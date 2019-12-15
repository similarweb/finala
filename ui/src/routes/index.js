import React from 'react'
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { Route, Switch } from 'react-router'
import Dashboard from '../components/Dashboard/Index'
import Resource from '../components/Resource/Index'
import Header from '../components/Header'
import LeftBar from '../components/LeftBar'
import NotFound from '../components/NotFound'
import { Col, Row } from "react-bootstrap";
import { ResourcesService } from "services/resources.service";

@connect()
export default class Routes extends React.Component {


  static propTypes = {    
    /**
     * Redux store
     */
    dispatch : PropTypes.func,
  
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
        // TODO:: show error when resources list is empty
      }
    );
  }

  render(){
    return(
      <div>
        <Header />
          <div id="main" className="container-fluid">
            
            <Row className="flex-xl-nowrap">
              <Col xl={2} sm={3} >
                <LeftBar/>
              </Col>
              <Col xl={10} sm={9} id="content-page">
                <Switch>
                  <Route exact path="/" component={Dashboard} />
                  <Route exact path="/resource/:name" component={Resource} />
                  <Route path="*" component={NotFound}/>
                </Switch>
                </Col>
              </Row>
            
            
          </div>
          
      </div>
    );
  }
}



