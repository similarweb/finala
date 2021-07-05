const initialState = {
  accounts: {},
};

/**
 * @param {object} state module state
 * @param {object} action to apply on state
 * @returns {object} new copy of state
 */
export function accounts(state = initialState, action) {
  switch (action.type) {
    case "ACCOUNT_LIST":
      state.accounts = action.data;
      return { ...state };
    default:
      return state;
  }
}
