import React from 'react'
import PropTypes from 'prop-types';
import SVG from 'react-inlinesvg';
import LoaderSVG from "styles/icons/loader-circle.svg"

/**
 * Application circle loader
 * Implementation example: <LoaderDots />
 */
export default class LoaderCircle extends React.Component {


  static propTypes = {    
    /**
     * Wrap load with className
     */
    wrapClass: PropTypes.string, 

    /**
     * 
     */
    bottomText: PropTypes.string, 
    };

  /**
   * Component render
   */    
  render() {
    return (
      <div className={this.props.wrapClass && this.props.wrapClass}>
        <SVG src={LoaderSVG} className="loader"/>
        {this.props.bottomText && <p>{this.props.bottomText}</p> }
      </div>
      )
  }
}



