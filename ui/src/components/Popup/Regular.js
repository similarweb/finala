import React from "react";
import PropTypes from "prop-types";
import { Modal, Button } from "react-bootstrap";
import ReactJson from 'react-json-view'

/**
 * Show log integration popup
 */
export default class PopupRegular extends React.Component {

  static propTypes = {      
    Show: PropTypes.func
  };

  state = {
    /**
     * Is popup is open
     */
    show: false,

    /**
     * Popup content
     */
    content: "",
  };

  /**
   * When component mount call the show popup handler
   */
  componentDidMount() {
    this.props.Show(this.handleShow);
  }

  /**
   * Close popup
   */
  handleClose = () => {
    this.setState({ show: false });
  }

  /**
   * Show popup
   */
  handleShow = (content) => {      
    this.setState({ show: true , content: content});
  }

  /**
  * Component render
  */    
  render() {
    if (this.state.content == "") {
      return (<div></div>);
    }
    let j = JSON.parse(this.state.content)
    return (
      <Modal size="lg" show={this.state.show} onHide={this.handleClose}>
        <Modal.Body>
          <pre>
            {
                  <div>
                    <ReactJson src={j} displayDataTypes={false} />
                  </div>
            }
          </pre>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.handleClose}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    );
  }
}
