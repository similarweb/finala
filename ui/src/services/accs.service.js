import { http } from "./request.service";

export const AccsService = {
  list,
};

/**
 *
 * @param {string} executionId execution to query
 */
function list(executionId) {
  return http
    .send(`api/v1/accounts/${executionId}`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}
