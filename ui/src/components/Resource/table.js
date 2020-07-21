import React from "react";
import {connect} from 'react-redux';
import PropTypes from 'prop-types';
import { ResourcesService } from "services/resources.service";
import TextUtils from "utils/Text";
import { withStyles } from '@material-ui/styles';
import { MuiThemeProvider, createMuiTheme } from "@material-ui/core/styles";
import MUIDataTable from "mui-datatables";
import numeral from 'numeral';
import TagsDialog from "../Dialog/Tags";
import LinearProgress from '@material-ui/core/LinearProgress';
import CircularProgress from '@material-ui/core/CircularProgress';
import EmojiPeopleIcon from '@material-ui/icons/EmojiPeople';
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';


const styles = () => ({
  icon:{
    fontSize: 80,
  }
});

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

@connect()
/**
 * Show resource data
 */
class Table extends React.Component {


 /**
  * Table option properties
  */
 options = {
  sizePerPageList: [ 5, 10, 15, 20 ],
  sizePerPage: 20,
  
  }
  
  static propTypes = {    
  /**
   * Match object
   */
    match: PropTypes.object,
    
    /**
     * Resource name
     */
    resources : PropTypes.object, 

    classes: PropTypes.object,

    resourceName: PropTypes.string,

    executionID: PropTypes.string

  };

  state = {

    resourceName: "",

    executionID: "",

    /**
     * Resource data
     */
    data : [],

    /**
     * Table headers
     */
    headers: [],

    /**
     * Fetch ajax timeout
     */
    timeoutAjaxCall: null,

    tableOptions: {
      selectableRows: false,
      responsive: "stacked",
    },

    showLoader: true,

  }
  
  /**
   * When component unmount, cancel the resource request
   */
  componentWillUnmount(){
    if (this.timeoutAjaxCall != null) {
         clearTimeout(this.timeoutAjaxCall);
    }
  }
  
  /**
   * When component updates, cancel the older request and start getting data with new resource name
   */
  componentDidUpdate(){

    if (this.state.resourceName != this.props.resourceName || this.props.executionID != this.state.executionID ){
      this.setState({showLoader: true })
      this.setState({resourceName: this.props.resourceName })
      this.setState({executionID: this.props.executionID })
      
      if (this.timeoutAjaxCall != null) {
        clearTimeout(this.timeoutAjaxCall);
      }
      this.getData(this.props.resourceName, this.props.executionID )
    }
    
  }

  /**
   * When component mount, getting data by resource name
   */
  componentDidMount(){

    this.setState({showLoader: true })
    this.setState({resourceName: this.props.resourceName })
    this.setState({executionID: this.props.executionID })
    this.getData(this.props.resourceName, this.props.executionID)
  }

  /**
   * Fetch data from the webserver
   * If the status of the resouce not not equal to 2, we still fetch the data for update without refresh
   * @param {string} resourceName 
   */
  getData(resourceName, executionID){
    
    ResourcesService.GetContent(resourceName, executionID).then(
      data => {
        if (data != null && typeof data == "object" && data.length > 0 ){
          
          const resources = []
          data.find(obj => {
            resources.push(obj.Data)
          })
          const firstRow = resources[0]
          const headers = []

          Object.keys(firstRow).map(function(key) {
            const header = {
              name: key,
              label: TextUtils.CamelCaseToTitleCase(key),
              options: {}
            }
            switch(key) {
              case "PricePerMonth":
              case "TotalSpendPrice":
                header["options"]["customBodyRender"] = (data) => {
                  return (
                  <span>{numeral(data).format('0,0[.]00 $')}</span>
                  )
                }
              break
              case "PricePerHour":
                header["options"]["customBodyRender"] = (data) => {
                  return (
                  <span>{numeral(data).format('0,0[.]00000 $')}</span>
                  )
                }
              break
              case "Tag":
                header["options"]["customBodyRender"] = (data) => {
                 return( <TagsDialog tags={data} />)
                }
                break
              default:
                header["options"]["customBodyRender"] = (data) => {
                  return (
                  <span>{data}</span>
                  )
                }

            }

            if (key !== "execution_id"){
              headers.push(header)
            }
            
          });

          this.setState({data:resources, headers})
          this.setState({showLoader: false })
        } else {
          this.setState({data: []})
          this.setState({showLoader: false })
        } 

        if (this.props.resources.Status != 2) {
          this.timeoutAjaxCall = setTimeout(() => { 
            this.getData(this.props.resourceName, executionID);
          }, 5000);
        }
      },
      () => {
        this.setState({data: []})
      }
    );
  }

  /**
  * Component render
  */  
  render() {
    return (
        <div id="resource-data" >
          {this.state.showLoader &&
            <Grid
            container
            spacing={0}
            direction="column"
            alignItems="center"
            justify="center"
            style={{ minHeight: '80vh', textAlign: "center" }}
            >
            <Grid item xs={3}>
              <CircularProgress size={50}/>
              <Typography variant="subtitle1" >
              Fetching data...
              </Typography>
            </Grid>   

            </Grid> 
          }
          { (!this.state.showLoader && this.props.resources.Status == 0 && this.state.data.length > 0 )&& <LinearProgress variant="query" />}
          { (!this.state.showLoader && this.props.resources.Status == 0 && this.state.data.length == 0 )&& 
            <Grid
            container
            spacing={0}
            direction="column"
            alignItems="center"
            justify="center"
            style={{ minHeight: '80vh', textAlign: "center" }}
            >

            <Grid item xs={3}>
              <CircularProgress size={50}/>
              <Typography variant="subtitle1" >
              Fetching data...
              </Typography>
            </Grid>   

            </Grid> 

          }
          { (!this.state.showLoader && this.props.resources.Status == 1 )&& 
              <Grid
              container
              spacing={0}
              direction="column"
              alignItems="center"
              justify="center"
              style={{ minHeight: '80vh', textAlign: "center" }}
              >

              <Grid item xs={10}>
                <ErrorOutlineIcon className={this.props.classes.icon}/>
                <Typography variant="subtitle1" >
                {this.props.resources.ErrorMessage}
                </Typography>
              </Grid>   

              </Grid> 
          
          }
          { (!this.state.showLoader && this.props.resources.Status == 2 && this.state.data.length == 0 )&& 

            <Grid
            container
            spacing={0}
            direction="column"
            alignItems="center"
            justify="center"
            style={{ minHeight: '80vh', textAlign: "center" }}
            >

            <Grid item xs={3}>
              <EmojiPeopleIcon className={this.props.classes.icon}/>
              <Typography variant="subtitle1" >
              Unused resources not found :)
              </Typography>
            </Grid>   

            </Grid> 
          }
           

          {!this.state.showLoader && this.state.data.length > 0 &&
            <MuiThemeProvider theme={getMuiTheme()}>
              <MUIDataTable
              title={TextUtils.ParseName(this.props.resourceName)}
              data={this.state.data}
              columns={this.state.headers}
              options={this.state.tableOptions}
              />
            </MuiThemeProvider>
            

          }
        </div>
    );
  }
}

export default withStyles(styles)(Table);
