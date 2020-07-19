import React, { useState, useEffect } from "react";
import { connect } from "react-redux";
import { Route, Switch } from "react-router";
import PropTypes from "prop-types";
import { ResourcesService } from "services/resources.service";
import { SettingsService } from "services/settings.service";
import Dashboard from "../components/Dashboard/Index";
import PageLoader from "../components/PageLoader";
import NotFound from "../components/NotFound";
import NoData from "../components/NoData";

import { CssBaseline, makeStyles, Box } from "@material-ui/core";

const useStyles = makeStyles(() => ({
  root: {
    background: "#f1f5f9",
    color: "#27303f",
    minHeight: "100vh",
  },
  content: {
    padding: "20px",
    background: "#f1f5f9",
    color: "#27303f",
  },
  hide: {
    display: "none",
  },
}));

let fetchTimeoutRequest = false;

/**
 * @param  {string} {currentExecution Global Execution Id
 * @param  {func} setCurrentExecution Update Current Execution
 * @param  {array} executions Executions list
 * @param  {func} setExecutions Update Executions list}
 */
const RouterIndex = ({
  currentExecution,
  setCurrentExecution,
  executions,
  setExecutions,
}) => {
  const classes = useStyles();
  const [isLoading, setIsLoading] = useState(true);

  /**
   * start fetching data from server
   */
  const init = () => {
    return SettingsService.GetSettings().then(
      () => {
        return fetchExecutions();
      },
      () => {}
    );
  };

  /**
   * fetch executions from server
   */

  const fetchExecutions = () => {
    clearTimeout(fetchTimeoutRequest);
    ResourcesService.GetExecutions()
      .then((responseData) => {
        const executions = responseData;
        setExecutions(executions);
        setIsLoading(false);
        if (executions.length) {
          const currentExecutionId = executions[0].ID;
          setCurrentExecution(currentExecutionId);
        } else {
          fetchTimeoutRequest = setTimeout(fetchExecutions, 5000);
        }
      })
      .catch(() => {
        fetchTimeoutRequest = setTimeout(fetchExecutions, 5000);
      });
  };

  /**
   * update state on execution change
   */
  useEffect(() => {
    if (!currentExecution) {
      init();
    } else {
      setIsLoading(false);
    }
  }, [currentExecution]);

  return (
    <div className={classes.root}>
      <CssBaseline />
      <main className={classes.content}>
        <Box component="div" m={3}>
          {isLoading && <PageLoader />}
          {!isLoading && !executions.length && <NoData />}
          {!isLoading && executions.length > 0 && (
            <Box component="div">
              <Switch>
                <Route exact path="/" component={Dashboard} />
                <Route path="*" component={NotFound} />
              </Switch>
            </Box>
          )}
        </Box>
      </main>
    </div>
  );
};

const mapStateToProps = (state) => ({
  currentExecution: state.executions.current,
  executions: state.executions.list,
});

const mapDispatchToProps = (dispatch) => ({
  setCurrentExecution: (id) => dispatch({ type: "EXECUTION_SELECTED", id }),
  setExecutions: (data) => dispatch({ type: "EXECUTION_LIST", data }),
});

RouterIndex.defaultProps = {};
RouterIndex.propTypes = {
  currentExecution: PropTypes.string,
  executions: PropTypes.array,
  setCurrentExecution: PropTypes.func,
  setExecutions: PropTypes.func,
};

export default connect(mapStateToProps, mapDispatchToProps)(RouterIndex);
