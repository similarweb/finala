const initialState = {
  list: [],
  current: "",
};

export function executions(state = initialState, action) {

  switch (action.type) {    
    case 'EXECUTION_SELECTED':  
      state.current = action.id
      return state;
    case 'EXECUTION_LIST':  
      state.list = action.data
      return state;
    default:
      return state
  }
}