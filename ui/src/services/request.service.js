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
   * @param {object} customRequestOptions customRequestOptions custom request options
   * @returns {Promise}
   */
  request(url, action, customRequestOptions = {}) {
    let defaultRequestOptions = {
      method: action,
      credentials: "include",
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

  /**
   * Making authentication request
   *
   * @param {string} url url request
   * @param {action} action method request (GET,POST,etc.)
   * @param {object} customRequestOptions customRequestOptions custom request options
   * @returns {Promise}
   */
  requestAuth(url, action, customRequestOptions) {
    let defaultRequestOptions = {
      method: action,
      credentials: "include",
    };
    merge(defaultRequestOptions, customRequestOptions);
    let fullUrl = "";
    if (url.startsWith("http")) {
      fullUrl = url;
    } else {
      fullUrl = `${this.baseURL}/${url}`;
    }
    return fetch(`${fullUrl}`, defaultRequestOptions).then(handleAuthResponse);
  }
}

/**
 * Manage http request response
 *
 * @param {response} url url request
 * @returns {Promise}
 */
function handleResponse(response) {
  return response.json().then((result) => {
    if (response.status == 200) {
      return result;
    }
    return Promise.reject(response);
  });
}

/**
 * Manage http login response
 *
 * @param {response} response request response
 * @returns {boolean, Promise} authentication correct or Promise, if rejected
 * */
function handleAuthResponse(response) {
  return response.text().then(() => {
    if (response.status === 200) {
      return true;
    } else {
      if (response.status === 401) {
        return false;
      }
    }
    return Promise.reject(response);
  });
}

const HTTPRequests = new Http();
export const http = {
  send: HTTPRequests.request,
  sendAuth: HTTPRequests.requestAuth,
};
