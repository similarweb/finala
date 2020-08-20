import React from "react";
import { connect } from "react-redux";
import { Route, Switch } from "react-router";
import PropTypes from "prop-types";
import Dashboard from "../components/Dashboard/Index";
import PageLoader from "../components/PageLoader";
import NotFound from "../components/NotFound";
import NoData from "../components/NoData";
import DataFactory from "../components/DataFactory";

import { CssBaseline, makeStyles, Box } from "@material-ui/core";

const useStyles = makeStyles(() => ({
  root: {
    background: "#f1f5f9",
    color: "#27303f",
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

/**
 * @param  {bool} isAppLoading App loading state
 * @param  {array} executions Executions list
 */
const RouterIndex = ({ isAppLoading, executions }) => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <CssBaseline />
      <DataFactory />
      <main className={classes.content}>
        <Box component="div" m={3}>
          {isAppLoading && <PageLoader />}
          {!isAppLoading && !executions.length && <NoData />}
          {!isAppLoading && executions.length > 0 && (
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
  executions: state.executions.list,
  isAppLoading: state.executions.isAppLoading,
});

const mapDispatchToProps = () => ({});

RouterIndex.defaultProps = {};
RouterIndex.propTypes = {
  isAppLoading: PropTypes.bool,
  executions: PropTypes.array,
  setCurrentExecution: PropTypes.func,
};

export default connect(mapStateToProps, mapDispatchToProps)(RouterIndex);
