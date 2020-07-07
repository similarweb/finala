import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from 'prop-types';
import {history} from 'configureStore'

import { TagsService } from "services/tags.service";

import { makeStyles } from '@material-ui/core/styles';
import {Box, Chip, TextField} from '@material-ui/core';
import Autocomplete from '@material-ui/lab/Autocomplete';



const useStyles = makeStyles((theme) => ({
   
  Autocomplete: {
    width: '100%',
  },
  filterInput: {
    borderColor:'#c1c1c1',
    backgroundColor: "white",
    '&:hover': {
      borderColor: 'red',
      borderWidth: 2
    },
  },
  chips: {
    fontWeight: 'bold',
    fontFamily:'Arial !important',
    margin: '5px',
    borderRadius: '3px',
    backgroundColor: '#d5dee6',
    fontSize: '14px'
  }


}));


const titleDirective = (title) => {
  let titleWords  = title.split('_').slice(1);
  titleWords = titleWords.map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
  return titleWords.join(' ');
} 

const FilterBar = ({ filters,currentExecution, setFilters, setResource }) => {

  const classes  = useStyles();
  const [tags, setTags] = useState({});
  const [options, setOptions] = useState([]);
  const [tagValues, setTagValues] = useState([]);
  let inputRef;


  const fetchTags = () => {
    TagsService.list(currentExecution).then(responseData => {
     

      const tagOptions = Object.keys(responseData).map(tagKey => {
        return { title: tagKey, id:tagKey}
      })
      setTags(responseData);
      setOptions(tagOptions)
      console.log('tags', tagOptions);
      
    })
  }

 
  const updateFilters = (filters) => {
    setFilters(filters);
    const searchParams = new window.URLSearchParams({filters: filters.map(f => f.id)})
    history.push({
      pathname: '/',
      search: `?${searchParams.toString()}`,
    });
  }

  const loadSearchState = () => {
    const searchParams = new window.URLSearchParams(window.location.search)
    const searchQuery = searchParams.get('filters');
    if (!searchQuery) {
      return;
    }
    const QFilters = searchQuery.split(',')
    if (QFilters[0] === "") {
      return;
    }
    let resource = false;
    const filters = [];
    QFilters.forEach(filter => {
        if (filter.substr(0,8) === 'resource') {
          let [filterKey, filterValue] = filter.split(':');
          const filterTitle = titleDirective(filterValue);
          filters.push({title: `Resource : ${filterTitle}`, id:filter, value: filterValue, type:'resource'})
          resource = filterValue;
        } else {
          const [filterKey, filterValue] = filter.split(' : ');
          filters.push({title: filter, id:filter, value: filterValue, type:'tag'})
        }
      
    });

    updateFilters(filters);
    if (resource) {
      setResource(resource)
    }
  }

  const optionChanged = (event, opt) => {
    if (!opt.length) {
      updateFilters([]);
      setResource(null);
      return;
    }
    if (opt.length < filters.length) {
      const filtersClone =  filters.slice(0, opt.length);
      updateFilters(filtersClone);

      const hasResourceFilter = filtersClone.findIndex(f => f.type === 'resource')
      if (hasResourceFilter === -1) {
        setResource(null);
      }
      return;
    }
    const id = opt[opt.length -1].id;
    filters.push({title: `${id} : `, id, type:'tag', value:null })

    updateFilters(filters);
    const tagValuesList =  tags[id].map(opt => {
      return  { title:`${id} : ${opt}`, id: `${id} : ${opt}`, value:opt, type:'tag'}
    })
    setTagValues(tagValuesList)
    inputRef.focus();
  }
  
  const onValueSelected = (event, opt) => {
    
    const filtersClone =  filters.slice(0, filters.length-1);
    const inFilters = filters.findIndex(row => row.id === opt.id)
    // prevent Duplicate
    if (inFilters === -1) {
      filtersClone.push({title: opt.title, id:opt.id, value: opt.value, type:'tag'})
    }
    updateFilters(filtersClone);
    setTagValues([]);
  }
  

  useEffect(() => {
    if (filters.length === 0) {
      loadSearchState();
    }
  },[filters]);

  useEffect(() => {
    if (currentExecution) {
      fetchTags();
    }
  },[currentExecution]);

  return (
    <Fragment>
      <Box mb={2}>
      <Autocomplete
    multiple
    value={filters}
    openOnFocus={true}
    className={classes.Autocomplete}
    id="fixed-tags-demo"
    onChange={optionChanged}
    getOptionSelected={(option, value) => false} 

    options={options}
    getOptionLabel={(option) => option.title}
    renderTags={(value, getTagProps) =>
      value.map((option) => (
        <Chip className={classes.chips} ma={2} label={option.title} key={option.title} />
      ))
    }
    renderInput={(params) => (
      <TextField {...params} className={classes.filterInput} variant="outlined" label="Add Filter" placeholder="Add Filter" />
    )}
  />
      <Autocomplete
    options={tagValues}
    onChange={onValueSelected}
    openOnFocus={true}
    getOptionLabel={(option) => option.title}
    getOptionSelected={(option, value) => false} 

    renderTags={(value, getTagProps) =>
      value.map((option) => (
        <Chip className={classes.chips}   ma={2} label={option.title} key={option.title} />
      ))
    }
    renderInput={(params) => (
      <TextField {...params}  inputRef={input => {
        inputRef = input;
      }} style={{visibility:'visible', marginTop:'-60px', zIndex:'-1'}} className={classes.filterInput} variant="outlined" label="" placeholder="" />
    )}
  />
      </Box>
    </Fragment>
    
  );
}

FilterBar.defaultProps = {};
FilterBar.propTypes = {
  filters: PropTypes.array,
  setFilters: PropTypes.func,
  setResource: PropTypes.func,
  currentExecution: PropTypes.string,
};



const mapStateToProps = state => ({
  filters: state.filters.filters,
  currentExecution: state.executions.current
});
const mapDispatchToProps = dispatch => ({
    setFilters: (data) =>  dispatch({ type: 'SET_FILTERS' , data}),
    setResource: (data) =>  dispatch({ type: 'SET_RESOURCE' , data})

});


// export default FilterBar;
export default connect(mapStateToProps, mapDispatchToProps)(FilterBar);
