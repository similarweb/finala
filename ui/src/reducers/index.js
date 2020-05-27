import { combineReducers } from 'redux'
import { connectRouter } from 'connected-react-router'
import { resources } from '../reducers/resources.reducer';
import { executions } from '../reducers/executions.reducer';

const rootReducer = (history) => combineReducers({
  resources,
  executions,
  router: connectRouter(history)
})

export default rootReducer
