import { http } from "./request.service";

export const AccsService = {
  fetchSummary,
};

/**
 *
 * @param {string} executionId execution to query
 */
function fetchSummary(executionId) {
  return http
    .send(`api/v1/getReport/${executionId}`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}
