import React from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import PageLoader from "../components/PageLoader";
import DataFactory from "../components/DataFactory";
import Login from "../components/Dashboard/Login";

import { Box, CssBaseline, makeStyles } from "@material-ui/core";
import NoData from "../components/NoData";
import {Dashboard} from "@material-ui/icons";
import NotFound from "../components/NotFound";
import {Route, Switch} from "react-router";

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
const RouterIndex = ({ isAppLoading, executions, authRequired }) => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <CssBaseline />
      <DataFactory />
      <main className={classes.content}>
        <Box component="div" m={3}>
          {authRequired && <Login />}
          {!authRequired && isAppLoading && <PageLoader />}
          {!authRequired && !isAppLoading && !executions.length && <NoData />}
          {!authRequired && !isAppLoading && executions.length > 0 && (
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
  authRequired: state.accounts.authRequired,
});

const mapDispatchToProps = () => ({});

RouterIndex.defaultProps = {};
RouterIndex.propTypes = {
  isAppLoading: PropTypes.bool,
  executions: PropTypes.array,
  setCurrentExecution: PropTypes.func,
  authRequired: PropTypes.bool,
};

export default connect(mapStateToProps, mapDispatchToProps)(RouterIndex);
