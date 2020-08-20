import { http } from "./request.service";

export const SettingsService = {
  GetSettings,
};

/**
 * fetch settings
 */
function GetSettings() {
  return http
    .send(`${window.location.origin}/api/v1/settings`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      http.baseURL = response.api_endpoint;
      return response;
    });
}
