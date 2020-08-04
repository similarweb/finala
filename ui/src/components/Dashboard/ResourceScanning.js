import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { titleDirective } from "../../directives";
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
 * @param  {func} setScanning  Update scanning status}
 */
const ResourceScanning = ({ resources, currentExecution, setScanning }) => {
  const classes = useStyles();
  const resource = Object.values(resources).find((row) => row.Status === 0);

  let isScanning = false;
  let title = "";

  if (resource) {
    isScanning = true;
    title = titleDirective(resource.ResourceName);
  } else {
    isScanning = false;
  }

  setScanning(isScanning);

  useEffect(() => {
    if (!currentExecution) {
      return;
    }
  }, [currentExecution, resources]);

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
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  currentExecution: state.executions.current,
});

const mapDispatchToProps = (dispatch) => ({
  setScanning: (isScanning) => dispatch({ type: "IS_SCANNING", isScanning }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceScanning);
