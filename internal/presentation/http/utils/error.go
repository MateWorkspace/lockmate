package presentationhttputils

import (
	"errors"
	"net/http"
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
)

type errorMapping struct {
	errType error
	status  int
	title   string
	message string
}

var errorMappings = []errorMapping{
	{domainmodelserror.NotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.BadArgs, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.Conflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.BadState, http.StatusPreconditionFailed, "Invalid State", ""},
	{domainmodelserror.Forbidden, http.StatusForbidden, "Access Denied", ""},
	{domainmodelserror.Unauthorized, http.StatusUnauthorized, "Unauthorized", "Authentication is required or no longer valid."},
	{domainmodelserror.Timeout, http.StatusGatewayTimeout, "Request Timeout", "The request took too long. Please try again."},
	{domainmodelserror.Unimplemented, http.StatusNotImplemented, "Not Implemented", "This feature is not available yet."},
	{domainmodelserror.Failure, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.Unknown, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.GoogleOAuthStateInvalid, http.StatusUnauthorized, "Unauthorized", "Authentication is required or no longer valid."},
	{domainmodelserror.GoogleOAuthDenied, http.StatusForbidden, "Access Denied", ""},
	{domainmodelserror.GoogleOAuthInvalid, http.StatusUnauthorized, "Unauthorized", "Authentication is required or no longer valid."},
	{domainmodelserror.GoogleOAuthFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.KeygenFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.KeyInvalid, http.StatusUnauthorized, "Unauthorized", "Authentication is required or no longer valid."},
	{domainmodelserror.KeyMismatch, http.StatusUnauthorized, "Unauthorized", "Authentication is required or no longer valid."},
	{domainmodelserror.PasswordMismatch, http.StatusBadRequest, "Invalid Password", "The current password is incorrect."},
	{domainmodelserror.PasswordHashFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.PasswordCompareFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.SpaceNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.SpaceSlugConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.PermissionNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.PermissionSlugConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.RoleNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.RoleSlugConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.RoleDefaultConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.DefaultRoleNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.RolePermissionNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.RolePermissionConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.UserNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.ClassicAuthNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.ClassicAuthUserConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.ClassicAuthUsernameConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.ClassicAuthEmailConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.ClassicAuthPhoneConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.GoogleAuthNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.GoogleAuthUserConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.GoogleAuthSubjectConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.ClientNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.ClientSlugConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.ClientOriginNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.ClientOriginConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.SpaceMemberNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.SpaceMemberConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.SpaceMemberRoleNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.SpaceMemberRoleConflict, http.StatusConflict, "Already Exists", ""},
	{domainmodelserror.TokenFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.TokenExpired, http.StatusUnauthorized, "Session Expired", "Your session has expired. Please sign in again."},
	{domainmodelserror.TokenInvalid, http.StatusUnauthorized, "Invalid Token", "Your session is no longer valid. Please sign in again."},
	{domainmodelserror.TraceNotFound, http.StatusNotFound, "Not Found", ""},
	{domainmodelserror.TraceEventIdFailed, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.PermissionSlugInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionSlugTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionSlugTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionDescInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.PermissionDescTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleSlugInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleSlugTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleSlugTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleDescInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.RoleDescTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserBioInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserBioTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserUsernameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserUsernameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserUsernameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPasswordInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPasswordTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPasswordTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserEmailInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserEmailTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserEmailTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPhoneInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPhoneTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserPhoneTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserApiKeyNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserApiKeyNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserApiKeyNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserApiKeyDescInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.UserApiKeyDescTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceSlugInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceSlugTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceSlugTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientSlugInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientSlugTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientSlugTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientNameInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientNameTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientNameTooShort, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientDescInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientDescTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientOriginInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.ClientOriginTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceDescInvalid, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.SpaceDescTooLong, http.StatusBadRequest, "Invalid Format", ""},
	{domainmodelserror.EnvironmentRequired, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."},
	{domainmodelserror.EnvironmentInvalid, http.StatusInternalServerError, "Internal Server Error", genericServerErrorMessage},
}

const genericServerErrorMessage = "Something went wrong on our end. Please try again later."

func Error(c *gin.Context, err error) {
	if err == nil {
		return
	}
	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.errType) {
			message := mapping.message
			if message == "" {
				message = domainMessage(err)
			}
			c.AbortWithStatusJSON(mapping.status, presentationhttpresponse.ErrorResponse{Error: mapping.title, Message: message})
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, presentationhttpresponse.ErrorResponse{Error: "Internal Server Error", Message: genericServerErrorMessage})
}

func domainMessage(err error) string {
	var value *domainmodelserror.Error
	if errors.As(err, &value) && strings.TrimSpace(value.Message) != "" {
		return value.Message
	}
	return genericServerErrorMessage
}

func MissingResponse(name string) error {
	return domainmodelserror.NewError(name+" is missing from usecase response", domainmodelserror.Failure, nil)
}
