import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { history } from "configureStore";
import { TagsService } from "services/tags.service";
import { makeStyles } from "@material-ui/core/styles";
import { Box, Chip, TextField } from "@material-ui/core";
import Autocomplete from "@material-ui/lab/Autocomplete";
import { titleDirective } from "../../directives";
import CancelIcon from "@material-ui/icons/Cancel";

const useStyles = makeStyles(() => ({
  Autocomplete: {
    width: "100%",
  },
  filterInput: {
    borderColor: "#c1c1c1",
    backgroundColor: "white",
    "&:hover": {
      borderColor: "red",
      borderWidth: 2,
    },
  },
  chips: {
    fontWeight: "bold",
    fontFamily: "Arial !important",
    margin: "5px",
    borderRadius: "3px",
    backgroundColor: "#d5dee6",
    fontSize: "14px",
  },
  valueAutoComplete: {
    visibility: "visible",
    marginTop: "-60px",
    zIndex: "-1",
  },
}));

/**
 * @param  {array} {filters  Filters List
 * @param  {string} currentExecution Global Execution Id
 * @param  {func} setFilters Update filters list
 * @param  {func} setResource Update Selected Resource}
 */
const FilterBar = ({ filters, currentExecution, setFilters, setResource }) => {
  const classes = useStyles();
  const [tags, setTags] = useState({});
  const [options, setOptions] = useState([]);
  const [tagValues, setTagValues] = useState([]);
  let inputRef;

  /**
   * Fetching server tagslist for autocomplete
   */
  const fetchTags = () => {
    TagsService.list(currentExecution).then((responseData) => {
      const tagOptions = Object.keys(responseData).map((tagKey) => {
        return { title: tagKey, id: tagKey };
      });
      setTags(responseData);
      setOptions(tagOptions);
    });
  };
  /**
   * Update filters list & history from auto complete
   * @param  {array} filters
   */
  const updateFilters = (filters) => {
    // verify uniqueness & has value
    const filtersList = filters.filter(
      (v, i, a) => a.findIndex((t) => t.id === v.id) === i
    );
    setFilters(filtersList);
    const searchParams = new window.URLSearchParams({
      filters: filters.map((f) => f.id),
    });
    history.push({
      pathname: "/",
      search: `?${searchParams.toString()}`,
    });
  };

  /**
   * Delete filter when X clicked
   * @param {object} filter Filter property from filter list
   */
  const deleteFilter = (filter) => {
    const updatedFilters = filters.filter((row) => row.id !== filter.id);
    updateFilters(updatedFilters);
    if (filter.type === "resource") {
      setResource(null);
    }
  };

  /**
   * Loading base state from url (tags and resource)
   */
  const loadSearchState = () => {
    const searchParams = new window.URLSearchParams(window.location.search);
    const searchQuery = searchParams.get("filters");
    if (!searchQuery) {
      return;
    }
    const QFilters = searchQuery.split(",");
    if (QFilters[0] === "") {
      return;
    }
    let resource = false;
    const filters = [];
    QFilters.forEach((filter) => {
      if (filter.substr(0, 8) === "resource") {
        let [, filterValue] = filter.split(":");
        if (filterValue) {
          const filterTitle = titleDirective(filterValue);
          filters.push({
            title: `Resource : ${filterTitle}`,
            id: filter,
            value: filterValue,
            type: "resource",
          });
          resource = filterValue;
        }
      } else {
        const [, filterValue] = filter.split(" : ");
        if (filterValue) {
          filters.push({
            title: filter,
            id: filter,
            value: filterValue,
            type: "tag",
          });
        }
      }
    });

    updateFilters(filters);
    if (resource) {
      setResource(resource);
    }
  };

  /**
   * Callback for Options select, read tag values and set options for 2nd autocomplete
   * @param {event} event Javascript onChange Event
   * @param {object} opt  Selected option from autocomplete
   */
  const optionChanged = (event, opt) => {
    if (!opt.length) {
      updateFilters([]);
      setResource(null);
      return;
    }
    if (opt.length < filters.length) {
      const filtersClone = filters.slice(0, opt.length);
      updateFilters(filtersClone);
      const hasResourceFilter = filtersClone.findIndex(
        (f) => f.type === "resource"
      );
      if (hasResourceFilter === -1) {
        setResource(null);
      }
      return;
    }
    // verify option is in options list
    if (!opt[opt.length - 1].id) {
      return false;
    }

    const id = opt[opt.length - 1].id;
    filters.push({ title: `${id} : `, id, type: "tag", value: null });

    updateFilters(filters);
    const tagValuesList = tags[id].map((opt) => {
      return {
        title: `${id} : ${opt}`,
        id: `${id} : ${opt}`,
        value: opt,
        type: "tag",
      };
    });
    setTagValues(tagValuesList);
    inputRef.focus();
  };

  /**
   * Callback for Tag Values autocomplete, adds the filter to the filters list
   * @param {event} event Javascript onChange Event
   * @param {object} opt  Selected option from autocomplete
   */
  const onValueSelected = (event, opt) => {
    const filtersClone = filters.slice(0, filters.length - 1);
    const inFilters = filters.findIndex((row) => row.id === opt.id);
    // prevent Duplicate
    if (inFilters === -1) {
      filtersClone.push({
        title: opt.title,
        id: opt.id,
        value: opt.value,
        type: "tag",
      });
    }
    updateFilters(filtersClone);
    setTagValues([]);
  };

  /**
   * Detect Autocomplete close reason, if its not because value selected, last selection will be removed
   * @param {event} event Javascript onChange Event
   * @param {string} opt  close type from autocomplete
   */
  const onValueClosed = (event, opt) => {
    if (opt !== "select-option") {
      const filtersClone = filters.slice(0, filters.length - 1);
      updateFilters(filtersClone);
      setTagValues([]);
    }
  };

  useEffect(() => {
    if (filters.length === 0) {
      loadSearchState();
    }
  }, [filters]);

  useEffect(() => {
    if (currentExecution) {
      fetchTags();
    }
  }, [currentExecution]);

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
          freeSolo
          options={options}
          getOptionLabel={(option) => option.title}
          getOptionSelected={() => false}
          renderTags={(value) =>
            value.map((option) => (
              <Chip
                className={classes.chips}
                ma={2}
                label={option.title}
                key={option.title}
                onDelete={() => deleteFilter(option)}
                deleteIcon={<CancelIcon />}
              />
            ))
          }
          renderInput={(params) => (
            <TextField
              {...params}
              className={classes.filterInput}
              variant="outlined"
              label="Add Filter"
              placeholder="Add Filter"
            />
          )}
        />
        <Autocomplete
          options={tagValues}
          onChange={onValueSelected}
          onClose={onValueClosed}
          openOnFocus={true}
          getOptionLabel={(option) => option.title}
          getOptionSelected={() => false}
          renderTags={(value) =>
            value.map((option) => (
              <Chip
                className={classes.chips}
                ma={2}
                label={option.title}
                key={option.title}
              />
            ))
          }
          renderInput={(params) => (
            <TextField
              {...params}
              inputRef={(input) => {
                inputRef = input;
              }}
              className={classes.valueAutoComplete}
              variant="outlined"
              label=""
              placeholder=""
            />
          )}
        />
      </Box>
    </Fragment>
  );
};

FilterBar.defaultProps = {};
FilterBar.propTypes = {
  filters: PropTypes.array,
  setFilters: PropTypes.func,
  setResource: PropTypes.func,
  currentExecution: PropTypes.string,
};

const mapStateToProps = (state) => ({
  filters: state.filters.filters,
  currentExecution: state.executions.current,
});
const mapDispatchToProps = (dispatch) => ({
  setFilters: (data) => dispatch({ type: "SET_FILTERS", data }),
  setResource: (data) => dispatch({ type: "SET_RESOURCE", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(FilterBar);
