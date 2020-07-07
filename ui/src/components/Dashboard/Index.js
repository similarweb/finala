import React, { Fragment } from "react";
import { connect } from "react-redux";
import { makeStyles } from '@material-ui/core/styles';

import PropTypes from 'prop-types';


import FilterBar from './FilterBar'
import StatisticsBar from './StatisticsBar'
import ResourcesChart from './ResourcesChart'
import ResourcesList from './ResourcesList'
import ResourceTable from './ResourceTable'
import ExecutionIndex from '../Executions/Index'

import { Grid, Box } from '@material-ui/core';


const useStyles = makeStyles((theme) => ({
  root: {
    width: '100%'
  },
  title: {
    fontFamily:'MuseoModerno'
  }

}));

  
const DashboardIndex = ({ currentResource }) => {

  const classes = useStyles();
  return (
    <Fragment>
      <Box mb={2}>
        <Grid container className={classes.root} spacing={2}>
          <Grid item sm={9} xs={12} style={{textAlign:'left'}}>
              <h1 className={classes.title}>Finala</h1>
          </Grid>
          <Grid item sm={3} xs={12} style={{textAlign:'right'}}>
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
}

DashboardIndex.defaultProps = {};
DashboardIndex.propTypes = {
  currentResource: PropTypes.string,
};


const mapStateToProps = state => ({
  currentResource: state.resources.currentResource,
});
const mapDispatchToProps = (dispatch) => ({});

export default connect(mapStateToProps, mapDispatchToProps)(DashboardIndex);