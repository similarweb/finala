import React, { Fragment } from "react";
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
 * @param  {array} {resources  Resources List}
 */
const ResourceScanning = ({ resources }) => {
  const classes = useStyles();
  const resource = Object.values(resources).find((row) => row.Status === 2);

  let isScanning = false;
  let title = "";

  if (resource) {
    isScanning = true;
    title = titleDirective(resource.ResourceName);
  }

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
  filters: PropTypes.array,
  addFilter: PropTypes.func,
  setResource: PropTypes.func,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
});

const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) => dispatch({ type: "ADD_FILTER", data }),
  setResource: (data) => dispatch({ type: "SET_RESOURCE", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceScanning);
