import React from "react";
import PropTypes from 'prop-types';
import {connect} from 'react-redux';
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
  }
 
});

@connect(state => ({
  resources: state.resources,
}))
/**
 * Application left bar menu
 */
class LeftBar extends React.Component {

  static propTypes = {    
    /**
     * List of all un-usage resources
     */
    resources : PropTypes.object, 

    classes: PropTypes.object

  };  

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
            {Object.keys(this.props.resources).map((resource) => (
              <ListItem button key={resource} component={Link} to={`/resource/${resource}`}>
                <ListItemText>
                <p className={this.props.classes.topLinkText}>{TextUtils.ParseName(resource)} ({this.props.resources[resource].ResourceCount})</p>
                <p className={this.props.classes.subLinkText}>{numeral(this.props.resources[resource].TotalSpent).format('0,0[.]00 $')}</p>
                </ListItemText>
                {this.props.resources[resource].Status == 1 && <ErrorOutlineIcon style={{position: "absolute", right: 5, top: 10, color: "red"}} />}
                {this.props.resources[resource].Status == 0 && <CircularProgress style={{position: "absolute", right: 5, top: 10}} className={this.props.classes.progress} size={16} />}
              </ListItem>
            ))}
          </List>
        </div>
      </Drawer>
    );
  }
}

export default withStyles(styles)(LeftBar);
