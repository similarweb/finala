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
import { SettingsService } from "services/settings.service";
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
  hide:{
    display: "none",
  }
});


@connect(state => ({
  selectedExecutionID: state.executions.current,
}))
class Routes extends React.Component {

  static propTypes = {    
    /**
     * Redux store
     */
    dispatch : PropTypes.func,
    
    classes: PropTypes.object,

    selectedExecutionID: PropTypes.string 
  
  };
    
  state = {
    /**
     * Fetch ajax timeout
     */
    timeoutAjaxCall: null,

    lastExecutionID: 0,

    executionsCount: null,
  }

  /**
   * When component mount, fetch resources data
   */
  componentDidMount() {
    SettingsService.GetSettings().then(() => {
        this.fetch()
      },
      () => {}
    );
  }

  /**
   * Fetch resources data
   */
  fetch(){
    ResourcesService.GetExecutions().then(
      data => {
        this.setState({executionsCount: data.length})
        if (data.length == 0){
          this.timeoutAjaxCall = setTimeout(() => { 
            this.fetch()
          }, 5000);
          return
        }
        this.props.dispatch({ type: 'EXECUTION_LIST', data})
        if (this.props.selectedExecutionID == ""){
          data.sort( this.compare );
          const lastExecution = data[0]
          this.setState({lastExecutionID: lastExecution.ID})
          this.props.dispatch({ type: 'EXECUTION_SELECTED', id: lastExecution.ID})
          this.timeoutAjaxCall = setTimeout(() => { 
            this.fetch()
          }, 5000);
        }
      },
      () => {
        this.timeoutAjaxCall = setTimeout(() => { 
          this.fetch()
        }, 5000);
      }
    );
  
  }

  compare( a, b ) {
    if ( a.Time > b.Time ){
      return -1;
    }
    if ( a.Time < b.Time ){
      return 1;
    }
    return 0;
  }

  render(){
    return(
      <div className={this.props.classes.root}>
        <CssBaseline />
        <Header />
        {this.state.lastExecutionID !== 0 && <LeftBar selectedExecutionID={this.state.lastExecutionID}/>}
        <main className={this.props.classes.content}>        
          <Toolbar />
          <Typography component={"div"}>
          <Box component="div" m={3}>
              {(this.state.executionsCount === null || this.state.executionsCount === 0) &&
                <Box component="div">
                    Waiting for the first collector run...
                </Box>
              }
              <Box component="div" className={(this.state.executionsCount === null || this.state.executionsCount === 0)  ? this.props.classes.hide : ""}>
                <Switch>
                  <Route exact path="/" component={Dashboard} />
                  <Route exact path="/resource/:name" component={Resource} />
                  <Route path="*" component={NotFound}/>
                </Switch>
              </Box>
            </Box>
            </Typography>
        </main>
      </div>
    );
  }
}

export default withStyles(styles)(Routes);
