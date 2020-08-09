const initialState = {
  filters: [],
};

/**
 * @param {object} state module state
 * @param {object} action to apply on state
 * @returns {object} new copy of state
 */
export function filters(state = initialState, action) {
  const pushFilter = (data) => {
    const id = action.data.id;

    if (data.type === "resource") {
      const inFiltersResourceIndex = state.filters.findIndex(
        (row) => row.type === "resource"
      );
      if (inFiltersResourceIndex !== -1) {
        state.filters[inFiltersResourceIndex] = action.data;
      }
    }

    const inFilters = state.filters.findIndex((row) => row.id === id);
    if (inFilters === -1) {
      state.filters.push(action.data);
    }

    state.filters = state.filters.filter(
      (row) => row.type !== "tag:incomplete"
    );

    return state.filters;
  };

  switch (action.type) {
    case "SET_FILTERS":
      state.filters = action.data;
      return { ...state };
    case "ADD_FILTER":
      state.filters = pushFilter(action.data).slice(0);
      return { ...state };
    default:
      return state;
  }
}
