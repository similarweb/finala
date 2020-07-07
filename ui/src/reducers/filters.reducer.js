const initialState = {
  filters: [],
};

export function filters(state = initialState, action) {


  const pushFilter = (data) => {
    const id = action.data.id;

    if (data.type === 'resource') {
      const inFiltersResourceIndex = state.filters.findIndex(row => row.type === 'resource');
      if (inFiltersResourceIndex !== -1) {
        state.filters[inFiltersResourceIndex] = action.data
      }
    }

    const inFilters = state.filters.findIndex(row => row.id === id);
    if (inFilters === -1) {
      // let filters = state.filters.slice(0); // create new copy
      state.filters.push(action.data);
      // return filters;
    }
    return state.filters;
  }

  switch (action.type) {    
    case 'SET_FILTERS':  
      state.filters = action.data
      // console.log(state)
      return {...state};
    case 'ADD_FILTER':  
      // console.log('ADD_FILTER dispatched' , new Date().getTime())
      state.filters = pushFilter(action.data).slice(0);
      return {...state};
    default:
      return state
    }
}