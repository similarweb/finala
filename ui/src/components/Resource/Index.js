import React from "react";
import {connect} from 'react-redux';
import PropTypes from 'prop-types';
import Table from './table'



@connect(state => ({
  resources: state.resources,
  selectedExecutionID: state.executions.current,
}))
/**
 * Show resource data
 */
class Resource extends React.Component {

  resourceName = this.props.match.params.name

  static propTypes = {    
    /**
     * Match object
     */
    match: PropTypes.object,
      
    /**
     * Resource name
     */
    resources : PropTypes.object, 

    /**
     * Current execution id
     */
    selectedExecutionID: PropTypes.string
 
  };

  getResource(){
    
    let resourceData = {}
   
    Object.keys(this.props.resources).map((resourceName) => {
      if (resourceName == this.props.match.params.name){
        resourceData = this.props.resources[resourceName]
        return
      }
    })

    return resourceData

  }

  /**
  * Component render
  */  
  render() {
    
    return (
        <div >
          {this.props.selectedExecutionID !== "" &&
            <Table 
              executionID={this.props.selectedExecutionID}
              resourceName={ this.props.match.params.name}
              resources={ this.getResource() }
            />
          }
        </div>
    );
  }
}

export default Resource;
