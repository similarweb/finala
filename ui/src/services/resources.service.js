import { http } from "./request.service";

export const ResourcesService = {
  GetExecutions,
  Summary,
  GetContent,
};

/**
 *
 * @param {array} filters filters list
 * @returns filters params for request
 */
const getTransformedFilters = (filters) => {
  const params = {};
  filters.forEach((filter) => {
    if (filter.id.substr(0, 8) === "resource") {
      return;
    }
    const [key, value] = filter.id.split(":");
    if (value) {
      const paramKey = `filter_Data.Tag.${key}`;
      if (params[paramKey]) {
        params[paramKey] += `,${value}`;
      } else {
        params[paramKey] = value;
      }
    }
  });
  return params;
};

/**
 * Get executions data
 */
function GetExecutions() {
  return http
    .send(`api/v1/executions`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}

/**
 *
 * @param {string} executionID execution to query
 * @param {array} filters filters list
 */
function Summary(executionID, filters = []) {
  const params = {
    ...getTransformedFilters(filters),
  };
  const searchParams = decodeURIComponent(
    new window.URLSearchParams(params).toString()
  );

  return http
    .send(`api/v1/summary/${executionID}?${searchParams}`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}

/**
 *
 * @param {string} name resource name
 * @param {string} executionID execution id to query
 * @param {array} filters filters list
 */
function GetContent(name, executionID, filters = []) {
  const params = {
    ...{ executionID },
    ...getTransformedFilters(filters),
  };
  const searchParams = new window.URLSearchParams(params).toString();

  return http
    .send(`api/v1/resources/${name}?${searchParams}`, `get`)
    .then(this.handleResponse)
    .then((response) => {
      return response;
    });
}
