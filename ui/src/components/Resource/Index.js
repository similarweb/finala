import React from "react";
import {connect} from 'react-redux';
import PropTypes from 'prop-types';
import { BootstrapTable, TableHeaderColumn } from 'react-bootstrap-table';
import { ResourcesService } from "services/resources.service";
import TextUtils from "utils/Text";
import NumberUtils from "utils/Number"
import LoaderCircle from "components/Loader/Circle";
import PopupRegular from "components/Popup/Regular"


@connect(state => ({
  resources: state.resources,
}))
/**
 * Show resouce data
 */
export default class Resource extends React.Component {

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
            headers.push( {
              id: key,
              text: TextUtils.ParseName(key),
              sort: true,
            })
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
   * Define the cell formant by field type
   * @param {string} type 
   */
  getCellFormat(type){    
    switch(type) {
      case "price_per_hour":
      case "price_per_month":
      case "total_spend_price":
        return this.priceFormatter
      case "tags":
        return this.popupFormatter
      default:
        return this.defaultFormatter
    }
  }

  /**
   * Cell content
   * @param {string} cell 
   */
  defaultFormatter(cell){
    return <p title={cell}>{cell}</p>

  }

  /**
   * Adding dolar char to the pricing cell
   * @param {float} cell 
   */
  priceFormatter(cell){
    return `$${NumberUtils.Format(cell,2)}`
  }

  /**
   * Shoe cell content in popup 
   * @param {string} cell 
   */
  popupFormatter(cell){
    return <p className="click" onClick={ () => this.ShowPopup(cell)}>Click to see tags</p>
  }

  /**
   * Show popup per cell
   */
  ShowPopup = (cell) => {
    this.clickChild(cell)
  }
  
  /**
  * Component render
  */  
  render() {
    const resourceFetchStatus = this.getStatus()
    return (
        <div id="resource-data">
        <h1>{TextUtils.ParseName(this.props.match.params.name)} <span>{(resourceFetchStatus == 0 && this.state.data.length > 0 ) && "(still in progress.. data will refresh automatically)"}</span></h1>
          
          { (resourceFetchStatus == 0 && this.state.data.length == 0 )&& <div className="center-loader"><LoaderCircle wrapClass="center" bottomText="Fetching data..." /></div>}
          { (resourceFetchStatus == 1 )&& <p>{this.getResourceDescription()}</p>}
          { (resourceFetchStatus == 2 && this.state.data.length == 0 )&& <p>Unused resources not found :)</p>}
          {this.state.data.length > 0 &&
            <BootstrapTable search pagination={ true }  options={ this.options } data={ this.state.data }>
            { this.state.headers.map((cel, index) => 
              <TableHeaderColumn dataFormat={ this.getCellFormat(cel.id).bind(this) } key={index} dataField={cel.id} isKey={index == 0} dataSort={ true }>{cel.text}</TableHeaderColumn>
            )}
            </BootstrapTable>
          }
          <PopupRegular Show={click => this.clickChild = click}/>
        </div>
    );
  }
}
