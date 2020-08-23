const initialState = {
  list: [],
  current: "",
  isScanning: false,
  isAppLoading: true,
};

/**
 * @param {object} state module state
 * @param {object} action to apply on state
 * @returns {object} new copy of state
 */
export function executions(state = initialState, action) {
  switch (action.type) {
    case "IS_APP_LOADING":
      state.isAppLoading = action.isLoading;
      return { ...state };
    case "IS_SCANNING":
      state.isScanning = action.isScanning;
      return { ...state };
    case "EXECUTION_SELECTED":
      state.current = action.id;
      return { ...state };
    case "EXECUTION_LIST":
      state.list = action.data;
      return { ...state };
    default:
      return state;
  }
}
