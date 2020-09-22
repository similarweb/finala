import React, { Fragment, useState, useEffect, useRef } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { setHistory, getHistory } from "../../utils/History";
import { TagsService } from "services/tags.service";
import { makeStyles } from "@material-ui/core/styles";
import { Box, Chip, TextField } from "@material-ui/core";
import Autocomplete from "@material-ui/lab/Autocomplete";
import { titleDirective } from "utils/Title";

let fetchTagsTimeout;
let debounceTimeout;
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
 * @param  {bool} isScanning indicate if the system is in scan mode
 * @param  {func} setFilters Update filters list
 * @param  {func} setResource Update Selected Resource}
 */
const FilterBar = ({
  filters,
  currentExecution,
  isScanning,
  setFilters,
  setResource,
}) => {
  const classes = useStyles();
  const [tags, setTags] = useState({});
  const [options, setOptions] = useState([]);
  const [defaultOptions, setDefaultOptions] = useState([]);
  const inputRef = useRef(null);
  const isScanningRef = useRef(isScanning);
  const isFilterBarOpen = useRef(false);

  /**
   * Fetching server tagslist for autocomplete
   */
  const fetchTags = async () => {
    clearTimeout(fetchTagsTimeout);
    const responseData = await TagsService.list(currentExecution).catch(
      () => false
    );
    if (!responseData) {
      fetchTagsTimeout = setTimeout(fetchTags, 5000);
      return false;
    }

    const tagOptions = Object.keys(responseData).map((tagKey) => ({
      title: tagKey.trim(),
      id: tagKey.trim(),
      type: "tag:option",
    }));

    if (!isFilterBarOpen.current) {
      setTags(responseData);
      setOptions(tagOptions);
    }
    setDefaultOptions(tagOptions);

    if (isScanningRef.current) {
      fetchTagsTimeout = setTimeout(fetchTags, 5000);
    }
  };
  /**
   * Update filters list & history from auto complete
   * @param  {array} filters
   */
  const updateFilters = (filters) => {
    setFilters(filters);
    setHistory({
      filters: filters,
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
    const searchQuery = getHistory("filters");
    if (!searchQuery) {
      return;
    }
    const QFilters = searchQuery.split(";");
    if (QFilters[0] === "") {
      return;
    }
    let resource = false;
    const filters = [];
    QFilters.forEach((filter) => {
      let [filterKey, filterValue] = filter.split(":");

      if (filterValue && filterKey === "resource") {
        const filterTitle = titleDirective(filterValue);
        filters.push({
          title: `Resource:${filterTitle}`,
          id: filter,
          value: filterValue,
          type: "resource",
        });
        resource = filterValue;
      } else if (filterValue) {
        const filterValues = filterValue.split(",");

        filterValues.forEach((filterValue) => {
          filters.push({
            title: `${filterKey}:${filterValue}`,
            id: `${filterKey}:${filterValue}`,
            type: "filter",
          });
        });
      }
    });

    setFilters(filters);
    if (resource) {
      setResource(resource);
    }
  };

  /**
   *
   * @param {any} opt -The value received from autocomplete-  Might be text or option value
   *  will detect if its a free text and find the real option as it was selected from the selectbox
   */
  const getOptionValueFromList = (opt) => {
    const lastOpt = opt[opt.length - 1];
    const isFreeText = !(lastOpt && lastOpt.id);
    let currentValue = opt[opt.length - 1];

    if (isFreeText) {
      const option = options.find((row) => {
        return (
          (row.type === "tag:option" && row.id === currentValue) ||
          (row.type === "tag:value" && row.title === currentValue)
        );
      });
      if (option) {
        currentValue = option;
        // if (option.type === "tag:option") {
        //   isTagOption = true;
        // }
      }
    }
    return currentValue;
  };

  /**
   *
   * @param {string} tagId the tag id from tags list
   * @returns {array} list of all values for selected tag
   */
  const getTagValueList = (tagId) => {
    const tagValuesList = tags[tagId].map((opt) => {
      return {
        title: `${opt}`,
        filterTitle: `${tagId}:${opt}`,
        id: `${tagId}:${opt}`,
        value: opt,
        type: "tag:value",
      };
    });
    return tagValuesList;
  };

  /**
   * Detect Autocomplete close reason, if its not because value selected, last selection will be removed
   * @param {event} event Javascript onChange Event
   * @param {string} opt  close type from autocomplete
   * FTI: its seems like blur is sent before select-option, debounce values to handle
   */
  const onValueClosed = (event, opt) => {
    clearTimeout(debounceTimeout);
    debounceTimeout = setTimeout(() => {
      if (opt !== "select-option" && opt !== "create-option") {
        //  remove incomplete tags
        isFilterBarOpen.current = false;
        filters = filters.filter((row) => row.type !== "tag:incomplete");
        updateFilters(filters);
        setOptions(defaultOptions); // reset options after selection
      }
    }, 50);
  };

  /**
   * Callback for Options select, read tag values and set options for 2nd autocomplete
   * @param {event} event Javascript onChange Event
   * @param {object} opt  Selected option from autocomplete
   */
  const optionChanged = (event, opt) => {
    isFilterBarOpen.current = false;
    // clear-all applied
    if (!opt.length) {
      updateFilters([]);
      setResource(null);
      setOptions(defaultOptions); // reset options after selection
      return;
    }
    // handle option delete (keyboard backspace)
    if (opt.length < filters.length) {
      filters = filters.slice(0, opt.length);
      updateFilters(filters);
      const hasResourceFilter = filters.findIndex((f) => f.type === "resource");
      if (hasResourceFilter === -1) {
        setResource(null);
      }
      setOptions(defaultOptions); // reset options after selection
      return;
    }

    const currentOption = getOptionValueFromList(opt);

    // not valid option
    if (!currentOption.id) {
      return false;
    }

    const isTagValue = currentOption && currentOption.type === "tag:value";
    const isTagOption = currentOption && currentOption.type === "tag:option";

    if (isTagOption) {
      // set new value lists
      const tagValuesList = getTagValueList(currentOption.id);
      setOptions(tagValuesList);
      // add filter that will be deleted later
      filters.push({
        title: `${currentOption.id}:`,
        id: currentOption.id,
        type: "tag:incomplete",
      });

      updateFilters(filters);
    }

    if (isTagValue) {
      filters.push({
        title: currentOption.filterTitle,
        id: currentOption.id,
        type: "filter",
      });

      // verify unique ids & remove incomplete tags
      filters = filters.filter((row, index) => {
        const filterFirstIndex = filters.findIndex((f) => f.id === row.id);
        return row.type !== "tag:incomplete" && filterFirstIndex === index;
      });

      updateFilters(filters);
      setOptions(defaultOptions); // reset options after selection
      return;
    }

    // trigger options open
    inputRef.current.blur();
    isFilterBarOpen.current = true;
    setTimeout(() => {
      inputRef.current.focus();
    });

    return;
  };

  /**
   * filters changed
   */
  useEffect(() => {
    if (filters.length === 0) {
      loadSearchState();
    }
    setOptions(defaultOptions); // reset options after selection
  }, [filters]);

  /**
   * currentExecution changed
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    isScanningRef.current = isScanning;
    fetchTags();
  }, [currentExecution]);

  /**
   * isScanning changed
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    if (isScanning !== isScanningRef.current) {
      isScanningRef.current = isScanning;
      if (isScanning) {
        fetchTagsTimeout = setTimeout(fetchTags, 5000);
      }
    }

    return () => {
      clearTimeout(fetchTagsTimeout);
    };
  }, [isScanning]);

  return (
    <Fragment>
      <Box mb={2}>
        <Autocomplete
          multiple
          value={filters}
          openOnFocus={true}
          className={classes.Autocomplete}
          onChange={optionChanged}
          onClose={onValueClosed}
          freeSolo
          options={options}
          getOptionLabel={(option) => option.title}
          getOptionSelected={() => false}
          renderTags={(value) =>
            value.map((option) => (
              <Fragment key={option.title}>
                {option.type === "tag:incomplete" && (
                  <span key={option.title}>{option.title}</span>
                )}
                {option.type !== "tag:incomplete" && (
                  <Chip
                    className={classes.chips}
                    ma={2}
                    label={option.title}
                    key={option.title}
                    onDelete={() => deleteFilter(option)}
                  />
                )}
              </Fragment>
            ))
          }
          renderInput={(params) => (
            <TextField
              {...params}
              className={classes.filterInput}
              inputRef={inputRef}
              variant="outlined"
              label="Add Filter"
              placeholder="Add Filter"
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
  isScanning: PropTypes.bool,
};

const mapStateToProps = (state) => ({
  filters: state.filters.filters,
  currentExecution: state.executions.current,
  isScanning: state.executions.isScanning,
});
const mapDispatchToProps = (dispatch) => ({
  setFilters: (data) => dispatch({ type: "SET_FILTERS", data }),
  setResource: (data) => dispatch({ type: "SET_RESOURCE", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(FilterBar);
