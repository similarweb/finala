import React from 'react'
import PropTypes from 'prop-types';
import LoaderSVG from "styles/icons/loader-dots.svg"
import SVG from 'react-inlinesvg';

/**
 * Render loader
 */
export default class LoaderDots extends React.Component {

  static propTypes = {    
    /**
     * Wrap load with className
     */
    wrapClass: PropTypes.string, 

    /**
     * Adding text in bottom of the loader
     */
    bottomText: PropTypes.string, 
    };

  /**
  * Component render
  */    
  render() {
    return (
      <div className={`${this.props.wrapClass} loader-dots`}>
        <SVG src={LoaderSVG} className="loader"/>
      </div>
        
        
      )
  }
}



