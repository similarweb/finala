import { http } from "./request.service";
export const AuthService = {
  Auth,
};

/**
 *
 * @param username {string} The users username
 * @param password {string} The users password
 */
function Auth(username, password) {
  const body = {
    Username: username,
    Password: password,
  };
  const customRequestOptions = {
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  };
  return http
    .send("api/v1/login", `get`, customRequestOptions)
    .then(this.handleAuthResponse)
    .then((response) => {
      return response;
    });
}
