package apiserver

import (
	"database/sql"
	"errors"
	"net/http"
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
	Data    *T     `json:"data, omitempty"`
	Message string `json:"message, omitempty"`
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

		if err := encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
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
