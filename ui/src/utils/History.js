import { history } from "configureStore";

const windowParams = new window.URLSearchParams(window.location.search);
let savedFilters = windowParams.get("filters");
let savedExecutionId = windowParams.get("executionId");
/**
 *
 * @param {filters, executionId} historyParams , filters: list of active filters, executionId: current selected execution (might be false for clearing property)
 *  Will set State into url search params, will save the params so we can set only part of the params
 */
export const setHistory = (historyParams = {}) => {
  savedFilters = historyParams.hasOwnProperty("filters")
    ? historyParams.filters
    : savedFilters;

  savedExecutionId = historyParams.hasOwnProperty("executionId")
    ? historyParams.executionId
    : savedExecutionId;
  const params = {};

  if (savedFilters && savedFilters.length) {
    params.filters = savedFilters;
  }
  if (savedExecutionId) {
    params.executionId = savedExecutionId;
  }

  const searchParams = new window.URLSearchParams(params);
  history.push({
    pathname: "/",
    search: decodeURIComponent(`?${searchParams.toString()}`),
  });
};

/**
 *
 * @param {string} query params name from url
 * @returns {string} Param value from url
 */
export const getHistory = (query) => {
  const searchParams = new window.URLSearchParams(window.location.search);
  const searchQuery = searchParams.get(query);
  return searchQuery;
};
