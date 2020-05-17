import React from "react";
import PropTypes from 'prop-types';
import {connect} from 'react-redux';
import { ResourcesService } from "services/resources.service";
import { withStyles } from '@material-ui/styles';
import Drawer from '@material-ui/core/Drawer';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import { Link } from "react-router-dom";
import CircularProgress from '@material-ui/core/CircularProgress';
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline';
import numeral from 'numeral';
import TextUtils from "utils/Text"
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import InputLabel from '@material-ui/core/InputLabel';
import Moment from 'moment';
import Grid from '@material-ui/core/Grid';
import Divider from '@material-ui/core/Divider';

const drawerWidth = 240;

const styles = () => ({
  drawer: {
    width: drawerWidth,
    flexShrink: 0,
  },
  drawerPaper: {
    width: drawerWidth,
  },
  drawerContainer: {
    overflow: 'auto',
  },
  progress:{
    // marginLeft: theme.spacing(2),
    marginRight: 4,
  },
  topLinkText:{
    marginBottom: 0,
    marginTop: 0,
  },
  subLinkText:{
    marginTop: 0,
    marginBottom: 0,
    color: "#939393",
    fontSize: 12,
  },
  executionSelect:{
    width: '100%'
  }
 
});

@connect(state => ({
  resources: state.resources,
  executions: state.executions,
}))
/**
 * Application left bar menu
 */
class LeftBar extends React.Component {

  static propTypes = {    
    resources : PropTypes.array, 
    dispatch : PropTypes.func,
    executions: PropTypes.object, 
    classes: PropTypes.object,
    selectedExecutionID: PropTypes.number

  };  

  state = {
    executionID: this.props.selectedExecutionID,

    /**
     * Fetch ajax timeout
     */
    timeoutAjaxCall: null,
  }

  componentDidMount() {
    this.fetch(this.state.executionID)
  }

  /**
   * Fetch resources data
   */
  fetch(executionID){
    ResourcesService.Summary(executionID).then(
        data => {
          this.props.dispatch({ type: 'RESOURCE_LIST', data: data[executionID]})
          this.timeoutAjaxCall = setTimeout(() => { 
            this.fetch(executionID)
          }, 5000);
        },
        () => {
          this.timeoutAjaxCall = setTimeout(() => { 
            this.fetch(executionID)
          }, 5000);
        }
      );
  }

  
  handleChange(event){
    
    this.props.dispatch({ type: 'EXECUTION_SELECTED', id: event.target.value})
    this.setState({executionID: event.target.value})
    clearTimeout(this.timeoutAjaxCall)
    this.fetch(event.target.value)
  }

  /**
  * Component render
  */    
  render() {
    return (
      <Drawer
        className={this.props.classes.drawer}
        variant="permanent"
        classes={{
          paper: this.props.classes.drawerPaper,
        }}
      >
        <Toolbar />
        <div className={this.props.classes.drawerContainer}>
        <List>
          <ListItem>
         
            <Grid container spacing={0}>
          <Grid item xs={12}>
            
              <InputLabel id="demo-simple-select-label">Executions</InputLabel>
              <Select
                className={this.props.classes.executionSelect}
                value={this.state.executionID}
                onChange={(event)=> this.handleChange(event)}
              >
                {this.props.executions.list.map((execution) => (
                  <MenuItem key={execution.ID} value={execution.ID}>{execution.ID} - {Moment(execution.CreatedAt).format('MM-DD-YYYY H:mm')}</MenuItem>
                ))}
              </Select>
              </Grid>
            </Grid>
            
          </ListItem>
        </List>
        <Divider />
        <List>
          {this.props.resources.map((resource, i) => (
                <ListItem button key={i} component={Link} to={`/resource/${resource.ResourceName}`}>
                <ListItemText>
                <p className={this.props.classes.topLinkText}>{TextUtils.ParseName(resource.ResourceName)} ({resource.ResourceCount})</p>
                <p className={this.props.classes.subLinkText}>{numeral(resource.TotalSpent).format('0,0[.]00 $')}</p>
                </ListItemText>
                {resource.Status == 1 && <ErrorOutlineIcon style={{position: "absolute", right: 5, top: 10, color: "red"}} />}
                {resource.Status == 0 && <CircularProgress style={{position: "absolute", right: 5, top: 10}} className={this.props.classes.progress} size={16} />}
                </ListItem>
          ))}
        </List>
        </div>
      </Drawer>
    );
  }
}

export default withStyles(styles)(LeftBar);
