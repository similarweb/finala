import { combineReducers } from 'redux'
import { connectRouter } from 'connected-react-router'
import { resources } from '../reducers/resources.reducer';

const rootReducer = (history) => combineReducers({
  resources,
  router: connectRouter(history)
})

export default rootReducer
