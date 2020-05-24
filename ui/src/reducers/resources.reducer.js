const initialState = {};

export function resources(state = initialState, action) {

  switch (action.type) {    
    case 'RESOURCE_LIST':  
      return action.data;
    default:
      return state
  }
}