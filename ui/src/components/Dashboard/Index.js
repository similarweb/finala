import React from "react";
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import TextUtils from "utils/Text"
import numeral from 'numeral';
import Highcharts from 'highcharts'
import HighchartsReact from 'highcharts-react-official'
import Paper from '@material-ui/core/Paper';
import MUIDataTable from "mui-datatables";
import { MuiThemeProvider, createMuiTheme } from "@material-ui/core/styles";
import CircularProgress from '@material-ui/core/CircularProgress';
import Typography from '@material-ui/core/Typography';


const tableFontSize = 12;

const getMuiTheme = () => createMuiTheme({
  overrides: {
    MUIDataTableHeadCell:{
      root:{
        color: "#878787"
      }
    },
    MUIDataTableBodyCell: {
      root: {
        fontSize: tableFontSize,
        
      },
      cellStackedSmall: { 
        fontSize: tableFontSize,
    },
    responsiveStackedSmall: { 
        fontSize: tableFontSize,
    },
      
    }
  }
});


/**
 * Dashboard page 
 */
@connect(state => ({
  resources: state.resources,
}))
export default class Dashboard extends React.Component {
  
  static propTypes = {    
    /**
     * List of all un-usage resources
     */
    resources : PropTypes.object, 
  };  

  state = {
    /**
     * Fetch ajax timeout
     */
    timeoutAjaxCall: null,

  }

  /**
   * Get PIE and Table data 
   */
  getData(){
    const pie = {
      chart: {
        plotBackgroundColor: null,
        plotBorderWidth: null,
        plotShadow: false,
        type: 'pie'
      },
      title: {
        text: 'Optional spend save'
      },
      tooltip: {
          pointFormat: '{series.name}: <b>{point.percentage:.1f}%</b>'
      },
      accessibility: {
          point: {
              valueSuffix: '%'
          }
      },
      plotOptions: {
          pie: {
              allowPointSelect: true,
              cursor: 'pointer',
              dataLabels: {
                  enabled: true,
                  format: '<b>{point.name}</b>: {point.percentage:.1f} %'
              }
          }
      },
        series: [
        
        ]
    };
    const table = {
      headers: [
        {label: "Resource".toUpperCase()}, 
        {label:"Optional spend save".toUpperCase(),options: { 
          sortDirection: 'desc',
          customBodyRender: (data) => {
            return (
            <span>{numeral(data).format('0,0[.]00 $')}</span>
            )
          }              
        }}],
      data:[]
    }
   
    const seriesData = []
    Object.keys(this.props.resources).map((resourceName) => {
      seriesData.push(
        {
          name: TextUtils.ParseName(resourceName),
          y: this.props.resources[resourceName].TotalSpent
      }
      )

      table.data.push(
        [TextUtils.ParseName(resourceName), this.props.resources[resourceName].TotalSpent]
      )

    })
    pie.series = [{data: seriesData}]
    return {pie, table}
  }

  /**
   * Component render
   */
  render() {
    const data = this.getData()
    return (
      <div className="">

        <h1>Dashboard</h1>
        {Object.keys(this.props.resources).length > 0 ?(
          <div>
            <Paper elevation={3} >
            <HighchartsReact
              allowChartUpdate={true}
                highcharts={Highcharts}
                options={data.pie}
            />
          </Paper>

          <br/>
          <MuiThemeProvider theme={getMuiTheme()}>
            <MUIDataTable
                  data={data.table.data}
                  columns={data.table.headers}
                  options={{selectableRows: false}}
              />
          </MuiThemeProvider>
          </div>
        ) : (
          <div
          style={{
              position: 'absolute', 
              left: '50%', 
              top: '50%',
              transform: 'translate(-50%, -50%)',
              textAlign: "center"
          }}
          >
          <CircularProgress size={50}/>
          <Typography variant="subtitle1" >
          Fetching data...
          </Typography>
          </div>
        )
        }

       
      </div>
    );
  }
}
