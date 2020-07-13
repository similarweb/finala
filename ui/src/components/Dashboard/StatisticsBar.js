import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { ResourcesService } from "services/resources.service";
import { titleDirective, MoneyDirective } from "../../directives";
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  Tooltip,
} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles(() => ({
  unused: {
    fontSize: "42px",
    color: "orangered",
    fontFamily: "MuseoModerno",
  },
  unused_daily: {
    fontSize: "42px",
    color: "purple",
    fontFamily: "MuseoModerno",
  },
  unused_resource: {
    fontSize: "42px",
    color: "darkgreen",
    fontFamily: "Nunito",
    fontWeight: "400",
  },
  middleGrid: {
    textAlign: "center",
    borderLeft: "1px dashed #c1c1c1",
    borderRight: "1px dashed #c1c1c1",
  },
  grid: {
    textAlign: "center",
  },
}));

/**
 * @param  {array} {resources  Resources List
 * @param  {array} filters  Filters List
 * @param  {func} currentResource  Current Selected Resource
 * @param  {func} currentExecution Current Selected Execution
 * @param  {func} setResources Update Resources List}
 */
const StatisticsBar = ({
  resources,
  filters,
  currentExecution,
  currentResource,
  setResources,
}) => {
  const classes = useStyles();

  let HighestResourceName = "";
  let HighestResourceValue = 0;
  const TotalSpent = Object.values(resources).reduce((acc, resource) => {
    let TotalSpent = resource.TotalSpent;

    if (currentResource && currentResource !== resource.ResourceName) {
      TotalSpent = 0;
    }

    if (resource.TotalSpent > HighestResourceValue) {
      HighestResourceValue = resource.TotalSpent;
      HighestResourceName = resource.ResourceName;
    }

    return acc + TotalSpent;
  }, 0);

  const DailySpent = TotalSpent / 30;

  /**
   * fetch Resources Summary
   */
  const getData = () => {
    ResourcesService.Summary(currentExecution, filters).then((responseData) => {
      setResources(responseData);
    });
  };

  /**
   * refetch data when filters or execution changes
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    getData();
  }, [filters, currentExecution]);

  return (
    <Fragment>
      <Box mb={3}>
        <Card>
          <CardContent>
            <Grid container className={classes.root} spacing={2}>
              <Grid item sm={4} xs={12} className={classes.grid}>
                <Tooltip title="Monthly Unused resources are effected from filters ">
                  <div>
                    <Typography className={classes.unused}>
                      {MoneyDirective(TotalSpent)}
                    </Typography>
                    <Typography>Monthly unused resources</Typography>
                  </div>
                </Tooltip>
              </Grid>
              <Grid item sm={4} xs={12} className={classes.middleGrid}>
                <Tooltip title="Daily waste is the amount you pay daily for unused resources and can be saved">
                  <div>
                    <Typography className={classes.unused_daily}>
                      {MoneyDirective(DailySpent)}
                    </Typography>
                    <Typography>Daily waste</Typography>
                  </div>
                </Tooltip>
              </Grid>
              <Grid item sm={4} xs={12} className={classes.grid}>
                <Typography className={classes.unused_resource}>
                  {titleDirective(HighestResourceName).toUpperCase()}
                </Typography>
                <Typography>Most unused resource</Typography>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      </Box>
    </Fragment>
  );
};

StatisticsBar.defaultProps = {};
StatisticsBar.propTypes = {
  currentExecution: PropTypes.string,
  currentResource: PropTypes.string,
  resources: PropTypes.object,
  filters: PropTypes.array,
  setResources: PropTypes.func,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
  currentExecution: state.executions.current,
  currentResource: state.resources.currentResource,
});
const mapDispatchToProps = (dispatch) => ({
  setResources: (data) => dispatch({ type: "RESOURCE_LIST", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(StatisticsBar);
