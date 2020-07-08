import { http } from "./request.service";

export const TagsService = {
  list,
};

/**
 *
 * @param {string} executionId execution to query
 */
function list(executionId) {
  return http
    .send(`api/v1/tags/${executionId}`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}
