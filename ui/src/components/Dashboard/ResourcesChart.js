import React, { Fragment } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import colors from "./colors.json";
import Chart from "react-apexcharts";
import { titleDirective } from "../../utils/Title";
import { MoneyDirective } from "../../utils/Money";
import { setHistory } from "../../utils/History";

import {
  Box,
  Card,
  CardContent,
  LinearProgress,
  makeStyles,
} from "@material-ui/core";
import ReportProblemIcon from "@material-ui/icons/ReportProblem";

const useStyles = makeStyles(() => ({
  noDataTitle: {
    textAlign: "center",
    fontWeight: "bold",
    margin: "5px",
    fontSize: "14px",
  },
  AlertIcon: {
    fontSize: "56px",
    color: "red",
  },
  progress: {
    margin: "30px",
  },
}));

/**
 * @param  {array} {resources  Resources List
 * @param  {array} filters  Filters List
 * @param  {bool} isResourceListLoading  isLoading state for resources
 * @param  {func} addFilter Add filter to  filters list
 * @param  {func} setResource Update Selected Resource}
 */
const ResourcesChart = ({
  resources,
  filters,
  isResourceListLoading,
  addFilter,
  setResource,
}) => {
  const classes = useStyles();
  const colorList = colors.map((color) => color.hex);
  const sortedResources = Object.values(resources)
    .filter((row) => row.TotalSpent > 0)
    .sort((a, b) => (a.TotalSpent >= b.TotalSpent ? -1 : 1));

  const chartOptions = {
    options: {
      chart: {
        type: "bar",
        width: "100%",
        events: {
          dataPointSelection: function (event, chartContext, config) {
            const dataPointIndex = config.dataPointIndex;
            const res = sortedResources;
            const selectedResource = res[dataPointIndex];
            setSelectedResource(selectedResource);
          },
        },
      },
      colors: colorList,
      tooltip: {
        theme: "light",
        x: {
          show: true,
        },
        y: {
          title: {
            formatter: function (val, opt) {
              return opt.w.globals.labels[opt.dataPointIndex];
            },
          },
        },
      },
      dataLabels: {
        enabled: true,
        formatter: function (val, opt) {
          return opt.w.globals.labels[opt.dataPointIndex];
        },
      },

      plotOptions: {
        bar: {
          horizontal: true,
          distributed: true,
          startingShape: "flat",
          endingShape: "flat",
          columnWidth: "70%",
          barHeight: "70%",
        },
      },
      xaxis: {
        categories: [],
      },
    },
    series: [
      {
        name: "",
        data: [],
      },
    ],
  };

  /**
   * update chart height according to number of resources
   */
  const getChartHeight = () => {
    if (!sortedResources || !sortedResources.length) {
      return 500;
    }
    return 120 * sortedResources.length;
  };
  /**
   * update chart height according to number of resources
   */
  const getCardHeight = () => {
    return getChartHeight() + 50;
  };

  /**
   *
   * @param {object} resource Set selected resource
   */
  const setSelectedResource = (resource) => {
    const filter = {
      title: `Resource:${resource.display_title}`,
      id: `resource:${resource.ResourceName}`,
      type: "resource",
    };
    setResource(resource.ResourceName);
    addFilter(filter);

    setHistory({
      filters: filters,
    });
  };

  /**
   * push resources into chart
   */
  sortedResources.forEach((resource) => {
    const title = titleDirective(resource.ResourceName);
    const amount = MoneyDirective(resource.TotalSpent);
    resource.title = `${title} (${amount})`;
    resource.display_title = `${title}`;

    chartOptions.options.xaxis.categories.push(resource.title);
    chartOptions.series[0].data.push(resource.TotalSpent);
    return resource;
  });

  return (
    <Fragment>
      <Box mb={3}>
        <Card>
          <CardContent style={{ minHeight: getCardHeight() }}>
            {!isResourceListLoading && sortedResources.length > 0 && (
              <Chart
                id="MainChart"
                height={getChartHeight()}
                options={chartOptions.options}
                series={chartOptions.series}
                type="bar"
              />
            )}
            {isResourceListLoading && (
              <LinearProgress className={classes.progress} />
            )}
            {!isResourceListLoading && !sortedResources.length && (
              <div className={classes.noDataTitle}>
                <ReportProblemIcon className={classes.AlertIcon} />
                <h3>No data found.</h3>
              </div>
            )}
          </CardContent>
        </Card>
      </Box>
    </Fragment>
  );
};

ResourcesChart.defaultProps = {};
ResourcesChart.propTypes = {
  resources: PropTypes.object,
  filters: PropTypes.array,
  isResourceListLoading: PropTypes.bool,
  addFilter: PropTypes.func,
  setResource: PropTypes.func,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  isResourceListLoading: state.resources.isResourceListLoading,
  filters: state.filters.filters,
});

const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) => dispatch({ type: "ADD_FILTER", data }),
  setResource: (data) => dispatch({ type: "SET_RESOURCE", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourcesChart);
