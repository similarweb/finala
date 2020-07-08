import React, { Fragment } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import colors from "./colors.json";
import Chart from "react-apexcharts";
import { titleDirective, MoneyDirective } from "../../directives";
import { history } from "configureStore";
import { Box, Card, CardContent } from "@material-ui/core";

/**
 * @param  {array} {resources  Resources List
 * @param  {array} filters  Filters List
 * @param  {func} addFilter Add filter to  filters list
 * @param  {func} setResource Update Selected Resource}
 */
const ResourcesChart = ({ resources, filters, addFilter, setResource }) => {
  const colorList = colors.map((color) => color.hex);
  const sortedResources = Object.values(resources).sort((a, b) =>
    a.TotalSpent >= b.TotalSpent ? -1 : 1
  );
  const chartOptions = {
    options: {
      chart: {
        type: "bar",
        width: "100%",
        height: "1500",
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
   *
   * @param {object} resource Set selected resource
   */
  const setSelectedResource = (resource) => {
    const filter = {
      title: `Resource : ${resource.title}`,
      id: `resource:${resource.ResourceName}`,
      type: "resource",
    };
    setResource(resource.ResourceName);
    addFilter(filter);

    const searchParams = new window.URLSearchParams({
      filters: filters.map((f) => f.id),
    });
    history.push({
      pathname: "/",
      search: `?${searchParams.toString()}`,
    });
  };

  /**
   * push resources into chart
   */
  sortedResources.forEach((resource) => {
    const title = titleDirective(resource.ResourceName);
    const amount = MoneyDirective(resource.TotalSpent);
    resource.title = `${title} (${amount})`;

    chartOptions.options.xaxis.categories.push(resource.title);
    chartOptions.series[0].data.push(resource.TotalSpent);
    return resource;
  });

  return (
    <Fragment>
      <Box mb={3}>
        <Card>
          <CardContent>
            <Chart
              id="MainChart"
              options={chartOptions.options}
              series={chartOptions.series}
              type="bar"
            />
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
  addFilter: PropTypes.func,
  setResource: PropTypes.func,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
});

const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) => dispatch({ type: "ADD_FILTER", data }),
  setResource: (data) => dispatch({ type: "SET_RESOURCE", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourcesChart);
