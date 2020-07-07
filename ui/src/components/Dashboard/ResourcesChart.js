import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from 'prop-types';
import colors from './colors.json'
import Chart from "react-apexcharts";

import { history } from 'configureStore'
import {
  Box,
  Card,
  CardContent
} from '@material-ui/core';
 

const titleDirective = (title) => {
  let titleWords = title.split('_').slice(1);
  titleWords = titleWords.map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
  return titleWords.join(' ');
}
const MoneyDirective = (amount, decimalCount = 2, decimal = ".", thousands = ",") => {
  try {
    decimalCount = Math.abs(decimalCount);
    decimalCount = isNaN(decimalCount) ? 2 : decimalCount;

    const negativeSign = amount < 0 ? "-" : "";

    let i = parseInt(amount = Math.abs(Number(amount) || 0).toFixed(decimalCount)).toString();
    let j = (i.length > 3) ? i.length % 3 : 0;

    return negativeSign + (j ? i.substr(0, j) + thousands : '') + i.substr(j).replace(/(\d{3})(?=\d)/g, "$1" + thousands) + (decimalCount ? decimal + Math.abs(amount - i).toFixed(decimalCount).slice(2) : "");
  } catch (e) {
    console.log(e)
  }
}

const ResourcesChart = ({
  resources,
  filters,
  addFilter,
  setResource
}) => {

  const colorList = colors.map(color => color.hex);
  const chartOptions = {
    options: {
      chart: {
        type: 'bar',
        width: '100%',
        height: '1500',
        events: {
          dataPointSelection: function (event, chartContext, config) {
            const seriesIndex = config.dataPointIndex;
            const res = Object.values(resources);
            const selectedResource = res[seriesIndex];
            setSelectedResource(selectedResource)
          },
        },
      },
      colors: colorList,
      tooltip: {
        theme: 'light',
        x: {
          show: true
        },
        y: {
          title: {
            formatter: function (val, opt) {
              return opt.w.globals.labels[opt.dataPointIndex];
            },
          }
        }
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
          startingShape: 'flat',
          endingShape: 'flat',
          columnWidth: '70%',
          barHeight: '70%',
        }
      },
      xaxis: {
        categories: []
      }
    },
    series: [{
      name: "",
      data: []
    }]
  };

  const setSelectedResource = (resource) => {
    const filter = {
      title: `Resource : ${resource.title}`,
      id: `resource:${resource.ResourceName}`,
      type: 'resource'
    }
    setResource(resource.ResourceName);
    addFilter(filter);
    
    const searchParams = new window.URLSearchParams({filters: filters.map(f => f.id)})
    history.push({
      pathname: '/',
      search: `?${searchParams.toString()}`,
    });
  }

  const resourcesList = Object.values(resources)
                        .sort((a, b) => (a.TotalSpent >= b.TotalSpent) ? -1 : 1)
                        .map(resource => {
                          const title = titleDirective(resource.ResourceName);
                          const amount = MoneyDirective(resource.TotalSpent);
                          resource.title = `${title} ($${amount})`;

                          chartOptions.options.xaxis.categories.push(resource.title);
                          chartOptions.series[0].data.push(resource.TotalSpent);
                          return resource;
                        });

  return ( <Fragment >
    <Box mb={3} >
    <Card>
    <CardContent >
      <Chart options={chartOptions.options}
            series={chartOptions.series}
            type = "bar"
            />
    </CardContent> 
    </Card>
    </Box>
    </Fragment>
  );
}

ResourcesChart.defaultProps = {};
ResourcesChart.propTypes = {
  resources: PropTypes.object,
  filters: PropTypes.array,
  addFilter: PropTypes.func,
  setResource: PropTypes.func,
};


const mapStateToProps = state => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
});

const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) =>  dispatch({ type: 'ADD_FILTER' , data}),
  setResource: (data) =>  dispatch({ type: 'SET_RESOURCE' , data})
});


export default connect(mapStateToProps, mapDispatchToProps)(ResourcesChart);