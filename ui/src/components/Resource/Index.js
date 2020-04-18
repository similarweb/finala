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

@connect(state => ({
  resources: state.resources,
}))
/**
 * Show resource data
 */
class Resource extends React.Component {

 resourceName = this.props.match.params.name

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

    classes: PropTypes.object

  };

  state = {
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

    }

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
    if (this.props.match.params.name != this.resourceName){
      this.resourceName = this.props.match.params.name
      
      if (this.timeoutAjaxCall != null) {
        clearTimeout(this.timeoutAjaxCall);
      }
      this.getData(this.props.match.params.name)
    }
    
  }

  /**
   * When component mount, getting data by resource name
   */
  componentDidMount(){
    this.getData(this.props.match.params.name)
  }

  /**
   * Fetch data from the webserver
   * If the status of the resouce not not equal to 2, we still fetch the data for update without refresh
   * @param {string} resourceName 
   */
  getData(resourceName){

    ResourcesService.GetContent(resourceName).then(
      data => {
        if (data != null && typeof data == "object" && data.length > 0 ){
          
          const firstRow = data[0]
          const headers = []

          Object.keys(firstRow).map(function(key) {
            const header = {
              name: key,
              label: TextUtils.ParseName(key).toUpperCase(),
              options: {}
            }
            switch(key) {
              case "price_per_month":
              case "total_spend_price":
                header["options"]["customBodyRender"] = (data) => {
                  return (
                  <span>{numeral(data).format('0,0[.]00 $')}</span>
                  )
                }
              break
              case "price_per_hour":
                header["options"]["customBodyRender"] = (data) => {
                  return (
                  <span>{numeral(data).format('0,0[.]00000 $')}</span>
                  )
                }
              break
              case "tags":
                header["options"]["customBodyRender"] = (data) => (
                  <TagsDialog tags={data} />
                  )
                break
            }
            headers.push(header)
          });

          this.setState({data, headers})
        } else {
          this.setState({data: []})
        } 

        if (this.getStatus() != 2) {
          this.timeoutAjaxCall = setTimeout(() => { 
            this.getData(this.props.match.params.name);
          }, 5000);
        }
      },
      () => {
        this.setState({data: []})
      }
    );
  }

  /**
   * Return the resource status.
   */
  getStatus(){
    const resourceData = this.props.resources[this.props.match.params.name]
    if (resourceData != undefined){
      return resourceData.Status
    }
    return null
    
  }

  getResourceDescription(){
    const resourceData = this.props.resources[this.props.match.params.name]
    if (resourceData != undefined){
      return resourceData.Description
    }
    return null
    
  }

  /**
  * Component render
  */  
  render() {
    const resourceFetchStatus = this.getStatus()
    return (
        <div id="resource-data">
          
          { (resourceFetchStatus == 0 && this.state.data.length > 0 )&& <LinearProgress variant="query" />}
          { (resourceFetchStatus == 0 && this.state.data.length == 0 )&& 
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
          { (resourceFetchStatus == 1 )&& 
              <Grid
              container
              spacing={0}
              direction="column"
              alignItems="center"
              justify="center"
              style={{ minHeight: '80vh', textAlign: "center" }}
              >

              <Grid item xs={3}>
                <ErrorOutlineIcon className={this.props.classes.icon}/>
                <Typography variant="subtitle1" >
                {this.getResourceDescription()}
                </Typography>
              </Grid>   

              </Grid> 
          
          }
          { (resourceFetchStatus == 2 && this.state.data.length == 0 )&& 

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
           

          {this.state.data.length > 0 &&
            <MuiThemeProvider theme={getMuiTheme()}>
              <MUIDataTable
              title={TextUtils.ParseName(this.props.match.params.name)}
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

export default withStyles(styles)(Resource);
