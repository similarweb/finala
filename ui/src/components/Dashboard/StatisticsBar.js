import React, { Fragment, useEffect, useState } from "react";
import { connect } from "react-redux";
import PropTypes from 'prop-types';
import numeral from 'numeral';

import { ResourcesService } from "services/resources.service";
import { titleDirective } from '../../directives'
import { Box, Card, CardContent, Grid, Typography } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles((theme) => ({
  unused: {
    fontSize:'42px', 
    color:'orangered', 
    fontFamily:'MuseoModerno'
  },
  unused_daily :{
    fontSize:'42px', 
    color:'purple', 
    fontFamily:'MuseoModerno'
  },
  unused_resource: {
    fontSize:'42px',
    color:'darkgreen', 
    fontFamily:'MuseoModerno'
  },
  middleGrid: {
    textAlign:'center', 
    borderLeft:'1px dashed #c1c1c1', 
    borderRight:'1px dashed #c1c1c1'
  },
  grid: {
    textAlign:'center', 
  },
}));
 
const StatisticsBar = ({ 
  resources,
  filters,
  currentExecution,
  setResources
  }) => {

  const classes = useStyles();

  let HighestResourceName = '';
  let HighestResourceValue = 0;
  const TotalSpent = Object.values(resources).reduce((acc, resource)=> {
    if (resource.TotalSpent > HighestResourceValue) {
      HighestResourceValue = resource.TotalSpent;
      HighestResourceName = resource.ResourceName;
    }
    return acc + resource.TotalSpent;
  },0)

  const DailySpent = TotalSpent/30;

  const getData = () => {
    ResourcesService.Summary(currentExecution, filters).then(responseData => {
      setResources(responseData);
    })
  };

  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    getData();
  }, [filters, currentExecution]);
 
  return (
    <Fragment>
      <Box mb={3}>
      <Card  >
      <CardContent>
      <Grid container className={classes.root} spacing={2}>
        <Grid item sm={4} xs={12} className={classes.grid}>
          <Typography className={classes.unused} >{numeral(TotalSpent).format('$0,0[.]00')}</Typography>
          <Typography   >Total unused resources</Typography>
        </Grid>
        <Grid item sm={4} xs={12} className={classes.middleGrid}>
          <Typography className={classes.unused_daily} >{numeral(DailySpent).format('$0,0[.]00')}</Typography>
          <Typography  >Daily waste</Typography>
        </Grid>
        <Grid item sm={4} xs={12} className={classes.grid}>
          <Typography className={classes.unused_resource} >{titleDirective(HighestResourceName)}</Typography>
          <Typography >Most unused resource</Typography>
        </Grid>
      </Grid>
      </CardContent>
      </Card>
      </Box>
    </Fragment>
    
  );
}

StatisticsBar.defaultProps = {};
StatisticsBar.propTypes = {
  currentExecution: PropTypes.string,
  resources: PropTypes.object,
  filters: PropTypes.array,
  setResources: PropTypes.func,
};


const mapStateToProps = state => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
  currentExecution: state.executions.current,

});
const mapDispatchToProps = (dispatch) => ({
  setResources: (data) =>  dispatch({ type: 'RESOURCE_LIST' , data})

});

export default connect(mapStateToProps, mapDispatchToProps)(StatisticsBar);