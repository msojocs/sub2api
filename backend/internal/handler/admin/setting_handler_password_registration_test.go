package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPasswordRegistrationSettingsRoundTrip(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{service.SettingKeyRegistrationEnabled: "true"})
	settings, err := h.settingService.GetAllSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.PasswordRegistrationEnabled, "existing installations keep password signup enabled")
	for _, enabled := range []bool{false, true, false} {
		rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true, "password_registration_enabled": enabled}, nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		settings, err = h.settingService.GetAllSettings(context.Background())
		require.NoError(t, err)
		require.Equal(t, enabled, settings.PasswordRegistrationEnabled)
		public, err := h.settingService.GetPublicSettings(context.Background())
		require.NoError(t, err)
		require.Equal(t, enabled, public.PasswordRegistrationEnabled)
		injected, err := h.settingService.GetPublicSettingsForInjection(context.Background())
		require.NoError(t, err)
		require.Equal(t, enabled, injected.(*service.PublicSettingsInjectionPayload).PasswordRegistrationEnabled)
	}
	rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "false", repo.values[service.SettingKeyPasswordRegistrationEnabled], "older clients must preserve the switch")
}
