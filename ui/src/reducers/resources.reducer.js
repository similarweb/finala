const initialState = {
  resources: {},
  currentResource: null,
  currentResourceData: [],
  isResourceListLoading: true,
  isResourceTableLoading: true,
};

/**
 * @param {object} state module state
 * @param {object} action to apply on state
 * @returns {object} new copy of state
 */
export function resources(state = initialState, action) {
  switch (action.type) {
    case "SET_CURRENT_RESOURCE_DATA":
      state.currentResourceData = action.data;
      return { ...state };
    case "IS_RESOURCE_TABLE_LOADING":
      state.isResourceTableLoading = action.isLoading;
      return { ...state };
    case "IS_RESOURCE_LIST_LOADING":
      state.isResourceListLoading = action.isLoading;
      return { ...state };
    case "CLEAR_RESOURCE":
      state.currentResource = null;
      return { ...state };
    case "SET_RESOURCE":
      state.currentResource = action.data;
      return { ...state };
    case "RESOURCE_LIST":
      state.resources = { ...action.data };
      return { ...state };
    default:
      return state;
  }
}
