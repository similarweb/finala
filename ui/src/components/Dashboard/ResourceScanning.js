import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { titleDirective } from "utils/Title";
import { Box, LinearProgress, makeStyles } from "@material-ui/core";

const useStyles = makeStyles(() => ({
  progress: {
    height: 2,
    maxWidth: 180,
    marginLeft: 20,
  },
  title: {
    textAlign: "left",
    marginTop: 5,
    marginLeft: 20,
    fontFamily: "MuseoModerno",
  },
}));

/**
 * will show a scanning message if some of the resources are still in progress
 * @param  {array} {resources  Resources List
 * @param  {string} currentExecution Current Selected Execution
 * @param  {bool} isScanning indicate if the system is in scan mode}
 */
const ResourceScanning = ({ resources, currentExecution, isScanning }) => {
  const classes = useStyles();
  const resource = Object.values(resources).find((row) => row.Status === 0);

  let title = "";
  if (resource) {
    title = titleDirective(resource.ResourceName);
  }

  /**
   * Re-render when currentExecution, resources, isScanning changes
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
  }, [currentExecution, resources, isScanning]);

  return (
    <Fragment>
      {isScanning && resource && (
        <Box mb={3}>
          {<LinearProgress className={classes.progress} />}
          <h3 className={classes.title}>Scanning: {title}</h3>
        </Box>
      )}
    </Fragment>
  );
};

ResourceScanning.defaultProps = {};
ResourceScanning.propTypes = {
  resources: PropTypes.object,
  currentExecution: PropTypes.string,
  setScanning: PropTypes.func,
  isScanning: PropTypes.bool,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  currentExecution: state.executions.current,
  isScanning: state.executions.isScanning,
});

const mapDispatchToProps = (dispatch) => ({
  setScanning: (isScanning) => dispatch({ type: "IS_SCANNING", isScanning }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceScanning);
