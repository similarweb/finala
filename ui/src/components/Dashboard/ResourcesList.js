import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from 'prop-types';
import colors from './colors.json'

import { makeStyles } from '@material-ui/core/styles';
import { history } from 'configureStore'
import { Box , Chip } from '@material-ui/core';
import { titleDirective, MoneyDirective } from '../../directives'

const useStyles = makeStyles((theme) => ({
  title: {
    fontFamily:'MuseoModerno'
  },
  resource_chips: {
    fontWeight: 'bold',
    fontFamily:'Arial !important',
    margin: '5px',
    borderRadius: '1px',
    backgroundColor: '#ffffff',
    borderLeft: '5px solid #ffffff',
    fontSize: '14px'
  }
}));

const ResourcesList = ({ 
  resources, 
  filters, 
  addFilter, 
  setResource }) => {

  const classes = useStyles();
  const resourcesList = Object.values(resources)
  .sort((a, b) => (a.TotalSpent > b.TotalSpent) ? -1 : 1)
  .map(resource => {
    const title = titleDirective(resource.ResourceName);
    const amount = MoneyDirective(resource.TotalSpent)
    resource.title = `${title} ($${amount})`
    return resource;
  });

  const setSelectedResource = (resource) => {
    const filter = {
      title: `Resource : ${resource.title}`, 
      id:`resource:${resource.ResourceName}`,
      type: 'resource'
    }
    setResource(resource.ResourceName);
    addFilter(filter);

    const searchParams = new window.URLSearchParams({filters: filters.map(f => f.id)})
    history.push({
      pathname: '/',
      search: `?${searchParams.toString()}`,
    });
  }

  return (
    <Fragment>
      <Box mb={3}>
      <h4 className={classes.title}>Resources:</h4>
      {resourcesList.map((resource, i) => <Chip className={classes.resource_chips} onClick={() => setSelectedResource(resource)} style={{borderLeftColor: colors[i].hex }}   ma={2} label={resource.title} key={i} />)}
      </Box>
    </Fragment>
  );
}

ResourcesList.defaultProps = {};
ResourcesList.propTypes = {
  resources: PropTypes.object,
  filters: PropTypes.array,
  addFilter: PropTypes.func,
  setResource: PropTypes.func,
};

const mapStateToProps = state => ({
  resources: state.resources.resources,
  filters: state.filters.filters,
});
const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) =>  dispatch({ type: 'ADD_FILTER' , data}),
  setResource: (data) =>  dispatch({ type: 'SET_RESOURCE' , data})
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourcesList);