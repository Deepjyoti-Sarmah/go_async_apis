package apiserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/reports"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignupRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		// var req SignupRequest
		// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 	return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
		// }
		// defer r.Body.Close()
		//
		// if err := req.Validate(); err != nil {
		// 	return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("invalid required: %v", err))
		// }

		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if existingUser != nil {
			return NewErrWithStatus(http.StatusConflict, err)
		}

		_, err = s.store.Users.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		// w.WriteHeader(http.StatusCreated)
		// if err := json.NewEncoder(w).Encode(ApiResponse[struct{}]{
		// 	Message: "successfully signed up user",
		// }); err != nil {
		// 	return NewErrWithStatus(http.StatusInternalServerError, err)
		// }

		if err := encode(ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SinginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r SigninRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (s *ApiServer) signinHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SigninRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		tokenPair, err := s.JwtManager.GenerateTokenPair(user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshTokenStore.Create(r.Context(), user.Id, tokenPair.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[SinginResponse]{
			Data: &SinginResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r TokenRefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

func (s *ApiServer) tokenRefreshHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[TokenRefreshRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		cureentRefreshToken, err := s.JwtManager.Parse(req.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userIdStr, err := cureentRefreshToken.Claims.GetSubject()
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		currentRefreshTokenRecord, err := s.store.RefreshTokenStore.ByPrimaryKey(r.Context(), userId, cureentRefreshToken)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}
			return NewErrWithStatus(status, err)
		}

		if currentRefreshTokenRecord.ExpiredAt.Before(time.Now()) {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("refresh_token expired"))
		}

		tokenPair, err := s.JwtManager.GenerateTokenPair(userId)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if _, err := s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), userId); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if _, err := s.store.RefreshTokenStore.Create(r.Context(), userId, tokenPair.RefreshToken); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[TokenRefreshResponse]{
			Data: &TokenRefreshResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type CreateReportRequest struct {
	ReportType string `json:"report_type"`
}

func (r CreateReportRequest) Validate() error {
	if r.ReportType == "" {
		return errors.New("report_type is required")
	}
	return nil
}

type ApiReport struct {
	Id                   uuid.UUID  `json:"id"`
	ReportType           string     `json:"report_type,omitempty"`
	OutputFilePath       *string    `json:"output_file_path,omitempty"`
	DownloadUrl          *string    `json:"download_url,omitempty"`
	DownloadUrlExpiresAt *time.Time `json:"download_url_expires_at,omitempty"`
	ErrorMessage         *string    `json:"error_message,omitempty"`
	CreatedAt            time.Time  `json:"created_at,omitempty"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	FailedAt             *time.Time `json:"failed_at,omitempty"`
	Status               string     `json:"status,omitempty"`
}

func (s *ApiServer) createReportHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[*CreateReportRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, ok := UserFromContext(r.Context())
		if !ok {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}
		report, err := s.store.ReportStore.Create(r.Context(), user.Id, req.ReportType)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		sqsMessage := reports.SqsMessage{
			UserId:   report.UserId,
			ReportId: report.Id,
		}

		bytes, err := json.Marshal(sqsMessage)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		queueUrlOutput, err := s.sqsClient.GetQueueUrl(r.Context(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(s.config.SqsQueue),
		})
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
			QueueUrl:    queueUrlOutput.QueueUrl,
			MessageBody: aws.String(string(bytes)),
		})
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[ApiReport]{
			Data: &ApiReport{
				Id:                   report.Id,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            report.StartedAt,
				CompletedAt:          report.CompletedAt,
				FailedAt:             report.FailedAt,
				Status:               report.Status(),
			},
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

func (s *ApiServer) getReportHander() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		reportIdStr := r.PathValue("id")
		reportId, err := uuid.Parse(reportIdStr)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, ok := UserFromContext(r.Context())
		if !ok {
			return NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("user not found in context"))
		}

		report, err := s.store.ReportStore.ByPrimaryKey(r.Context(), user.Id, reportId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return NewErrWithStatus(http.StatusNotFound, err)
			}
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if report.CompletedAt != nil {
			needsRefresh := report.DownloadUrlExpiresAt != nil && report.DownloadUrlExpiresAt.Before(time.Now())
			// to s3 client (presigned client)
			if report.DownloadUrl == nil || needsRefresh {

				expiresAt := time.Now().Add(time.Second * 10)
				signedUrl, err := s.presignedClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
					Bucket: aws.String(s.config.S3Bucket),
					Key:    report.OutputFilePath,
				}, func(options *s3.PresignOptions) {
					options.Expires = time.Second * 10
				})
				if err != nil {
					return NewErrWithStatus(http.StatusInternalServerError, err)
				}
				report.DownloadUrl = &signedUrl.URL
				report.DownloadUrlExpiresAt = &expiresAt

				report, err = s.store.ReportStore.Update(r.Context(), report)
				if err != nil {
					return NewErrWithStatus(http.StatusInternalServerError, err)
				}
			}
		}

		if err := encode(ApiResponse[ApiReport]{
			Data: &ApiReport{
				Id:                   report.Id,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            report.StartedAt,
				CompletedAt:          report.CompletedAt,
				FailedAt:             report.FailedAt,
				Status:               report.Status(),
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
