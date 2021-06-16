import { combineReducers } from "redux";
import { connectRouter } from "connected-react-router";
import { accounts } from "../reducers/accounts.reducer";
import { resources } from "../reducers/resources.reducer";
import { executions } from "../reducers/executions.reducer";
import { filters } from "../reducers/filters.reducer";

const rootReducer = (history) =>
  combineReducers({
    accounts,
    resources,
    executions,
    filters,
    router: connectRouter(history),
  });

export default rootReducer;
