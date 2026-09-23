package service

import (
	"context"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// LoginOrRegisterOIDCWithSignupCodes creates passwordless OIDC accounts. Unverified
// upstream emails must never be used as a local account's email or ownership proof.
func (s *AuthService) LoginOrRegisterOIDCWithSignupCodes(ctx context.Context, input EmailOAuthIdentityInput, invitationCode, affiliateCode, promoCode string) (*TokenPair, *User, error) {
	if input.ProviderType != "oidc" || strings.TrimSpace(input.ProviderKey) == "" || strings.TrimSpace(input.ProviderSubject) == "" {
		return nil, nil, infraerrors.BadRequest("OAUTH_IDENTITY_INVALID", "invalid oidc identity")
	}
	metadata := make(map[string]any, len(input.UpstreamMetadata)+2)
	for key, value := range input.UpstreamMetadata {
		metadata[key] = value
	}
	if syntheticEmail, ok := metadata["email"].(string); ok && strings.HasSuffix(syntheticEmail, OIDCConnectSyntheticEmailDomain) && syntheticEmail != input.Email {
		metadata["synthetic_email"] = syntheticEmail
	}
	metadata["email"] = input.Email
	metadata["email_verified"] = input.EmailVerified
	input.UpstreamMetadata = metadata
	if input.EmailVerified {
		return s.LoginOrRegisterVerifiedEmailOAuthWithSignupCodes(ctx, input, invitationCode, affiliateCode, promoCode)
	}
	if !strings.HasSuffix(input.Email, OIDCConnectSyntheticEmailDomain) {
		return nil, nil, infraerrors.BadRequest("OAUTH_EMAIL_NOT_VERIFIED", "unverified oidc email requires a synthetic account address")
	}
	if s == nil || s.entClient == nil {
		return nil, nil, ErrServiceUnavailable
	}
	tokenPair, user, err := s.LoginOrRegisterOAuthWithTokenPairAndPromoCode(ctx, input.Email, input.Username, invitationCode, affiliateCode, promoCode, "oidc")
	if err != nil {
		return nil, nil, err
	}
	if err := s.ensureEmailOAuthIdentity(ctx, user.ID, input); err != nil {
		return nil, nil, err
	}
	s.RecordSuccessfulLogin(ctx, user.ID)
	return tokenPair, user, nil
}
