import React, { Fragment } from "react";
import { connect } from "react-redux";
import { Link } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import PropTypes from "prop-types";
import FilterBar from "./FilterBar";
import StatisticsBar from "./StatisticsBar";
import ResourceScanning from "./ResourceScanning";
import ResourcesChart from "./ResourcesChart";
import ResourcesList from "./ResourcesList";
import ResourceTable from "./ResourceTable";
import ExecutionIndex from "../Executions/Index";
import Logo from "../Logo";
import { Grid, Box } from "@material-ui/core";

const useStyles = makeStyles(() => ({
  root: {
    width: "100%",
  },
  title: {
    fontFamily: "MuseoModerno",
  },
  logoGrid: {
    textAlign: "left",
  },
  selectorGrid: {
    textAlign: "right",
  },
}));

/**
 * @param  {string} {currentResource  Current Selected Resource}
 */
const DashboardIndex = ({ currentResource }) => {
  const classes = useStyles();
  return (
    <Fragment>
      <Box mb={2}>
        <Grid container className={classes.root} spacing={0}>
          <Grid item sm={9} xs={12} className={classes.logoGrid}>
            <Link to="/">
              <Logo />
            </Link>
            <ResourceScanning />
          </Grid>
          <Grid item sm={3} xs={12} className={classes.selectorGrid}>
            <ExecutionIndex />
          </Grid>
        </Grid>
      </Box>

      <FilterBar />
      <StatisticsBar />
      {!currentResource && <ResourcesChart />}
      {currentResource && <ResourcesList />}
      {currentResource && <ResourceTable />}
    </Fragment>
  );
};

DashboardIndex.defaultProps = {};
DashboardIndex.propTypes = {
  currentResource: PropTypes.string,
};

const mapStateToProps = (state) => ({
  currentResource: state.resources.currentResource,
});
const mapDispatchToProps = () => ({});

export default connect(mapStateToProps, mapDispatchToProps)(DashboardIndex);
