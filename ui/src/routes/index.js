import React from 'react'
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { Route, Switch } from 'react-router'
import Dashboard from '../components/Dashboard/Index'
import Resource from '../components/Resource/Index'
import Header from '../components/Header'
import LeftBar from '../components/LeftBar'
import NotFound from '../components/NotFound'
import { ResourcesService } from "services/resources.service";
import Typography from '@material-ui/core/Typography';
import { withStyles } from '@material-ui/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Toolbar from '@material-ui/core/Toolbar';
import Box from '@material-ui/core/Box';


const styles = () => ({
  root: {
    display: 'flex',
  },
  content: {
    flexGrow: 1,
    padding: 3,
  },
});


@connect()
class Routes extends React.Component {

  static propTypes = {    
    /**
     * Redux store
     */
    dispatch : PropTypes.func,
    
    classes: PropTypes.object
  
  };
    
  state = {
    /**
     * Fetch ajax timeout
     */
    timeoutAjaxCall: null,
  }

  /**
   * When component mount, fetch resources data
   */
  componentDidMount() {
    this.fetch()
  }

  /**
   * Fetch resources data
   */
  fetch(){
    ResourcesService.Summary().then(
      data => {
        this.props.dispatch({ type: 'RESOURCE_LIST', data})
        this.timeoutAjaxCall = setTimeout(() => { 
          this.fetch()
        }, 5000);
      },
      () => {
        this.timeoutAjaxCall = setTimeout(() => { 
          this.fetch()
        }, 5000);
      }
    );
  }

  render(){
    return(
      <div className={this.props.classes.root}>
        <CssBaseline />
        <Header />
        <LeftBar/>
        <main className={this.props.classes.content}>
          <Toolbar />
          <Typography component={"div"}>
            <Box component="div" m={3}>
              <Switch>
                <Route exact path="/" component={Dashboard} />
                <Route exact path="/resource/:name" component={Resource} />
                <Route path="*" component={NotFound}/>
              </Switch>
            </Box>
            </Typography>
          
        </main>
       
       
        
                
          
      </div>
    );
  }
}

export default withStyles(styles)(Routes);
