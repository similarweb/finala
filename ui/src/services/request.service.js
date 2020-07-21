var merge = require("lodash/merge");

/**
 *  Manage http request
 */
class Http {
  baseURL = "";
  /**
   * Making request
   *
   * @param {string} url url request
   * @param {action} action method request (GET,POST,etc.)
   * @param {object} url customRequestOptions custom request options
   * @returns {Promise}
   */
  request(url, action, customRequestOptions = {}) {
    let defaultRequestOptions = {
      method: action,
    };
    merge(defaultRequestOptions, customRequestOptions);
    let fullUrl = "";
    if (url.startsWith("http")) {
      fullUrl = url;
    } else {
      fullUrl = `${this.baseURL}/${url}`;
    }

    return fetch(`${fullUrl}`, defaultRequestOptions).then(handleResponse);
  }
}

/**
 * Manage http request response
 *
 * @param {response} url url request
 * @returns {object}
 */
function handleResponse(response) {
  return response.json().then((result) => {
    if (response.status == 200) {
      return result;
    }
    return Promise.reject(response);
  });
}

const HTTPRequests = new Http();
export const http = {
  send: HTTPRequests.request,
};
