const initialState = {
  authRequired: true,
};

/**
 * @param {object} state module state
 * @param {object} action to apply on state
 * @returns {object} new copy of state
 */
export function auth(state = initialState, action) {
  switch (action.type) {
    case "SET_AUTH_REQUIRED":
      state.authRequired = action.data;
      return { ...state };
    default:
      return state;
  }
}
