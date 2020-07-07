import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import { Route, Switch } from 'react-router'

import PropTypes from 'prop-types';


import { ResourcesService } from "services/resources.service";
import { SettingsService } from "services/settings.service";

import Dashboard from '../components/Dashboard/Index'
import PageLoader from '../components/PageLoader'
import NotFound from '../components/NotFound'

import {
    CssBaseline,
    makeStyles,
    Box
} from '@material-ui/core';
 
const useStyles = makeStyles((theme) => ({
    root: {
        background:'#f1f5f9',
        color: '#27303f'
    },
    content: {
    padding: '20px',
    background:'#f1f5f9',
    color: '#27303f'
    },
    hide:{
    display: "none",
    }
    
}));
  

const RouterIndex = ({ 
    currentExecution,
    setCurrentExecution,
    executions,
    setExecutions,

 }) => {

    const classes = useStyles();
    const [isLoading, setIsLoading] = useState(true);

    const init =  () => {
        return SettingsService.GetSettings().then(() => {
            return fetchExecutions()
          },
          () => {}
        );
    }

    const fetchExecutions = () => {
        ResourcesService.GetExecutions().then(responseData => {
            const executions = responseData.reverse(); // executions list
            setExecutions(executions);
            setIsLoading(false);
            if (executions.length) {
                const currentExecutionId = executions[0].ID;
                setCurrentExecution(currentExecutionId);
            }  
        })
    };

    useEffect(() => {
        if (!currentExecution) {
            init();
        } else {
            setIsLoading(false);
        }
      },[currentExecution]);

    return(
        <div className={classes.root}>
          <CssBaseline />
          <main className={classes.content}>        
            <Box component="div" m={3}>
                {isLoading && <PageLoader/>}
                {!isLoading && !executions.length &&
                  <Box component="div">
                      Waiting for the first collection of data for Finala
                  </Box>
                }
                {!isLoading && executions.length && 
                 <Box component="div">
                  <Switch>
                    <Route exact path="/" component={Dashboard} />
                    <Route path="*" component={NotFound}/>
                  </Switch>
                </Box> 
                }
              </Box>
          </main>
        </div>
      );
};




const mapStateToProps = state => ({
    currentExecution: state.executions.current,
    executions: state.executions.list
  });
  
const mapDispatchToProps = (dispatch) => ({
setCurrentExecution: (data) =>  dispatch({ type: 'EXECUTION_SELECTED' , id: data }),
setExecutions: (data) =>  dispatch({ type: 'EXECUTION_LIST' , data }),
});


 

RouterIndex.defaultProps = {};
RouterIndex.propTypes = {
  currentExecution: PropTypes.string,
  executions: PropTypes.array,
  setCurrentExecution: PropTypes.func,
  setExecutions: PropTypes.func,
};
  
export default connect(mapStateToProps, mapDispatchToProps)(RouterIndex);