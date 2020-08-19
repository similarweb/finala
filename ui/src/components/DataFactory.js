import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { ResourcesService } from "services/resources.service";
import { SettingsService } from "services/settings.service";
import { titleDirective } from "../directives";
import { getHistory, setHistory } from "../utils/History";

let fetchTimeoutRequest = false;
let initTimeoutRequest = false;
let lastFiltersSearched = "";

/**
 * will show a scanning message if some of the resources are still in progress
 * {
 * @param  {func} setExecutions Update Executions list
 * @param  {string} currentExecution Current Selected Execution
 * @param  {func} setCurrentExecution Update Current Execution
 *
 * @param  {func} setResources Update Resources List
 * @param  {func} setIsResourceListLoading  update isLoading state for resources
 *
 * @param  {array} resources  Resources List
 * @param  {bool} isScanning indicate if the system is in scan mode
 * @param  {array} filters  Filters List
 *
 * @param  {func} setIsAppLoading  Update App IsLoading status
 * @param  {func} setIsScanning  Update scanning status
 * }
 */
const DataFacotry = ({
  setExecutions,
  currentExecution,
  setCurrentExecution,

  setResources,
  setIsResourceListLoading,

  filters,
  setIsAppLoading,
  setIsScanning,
}) => {
  /**
   * start fetching data from server
   * will load executions list
   */
  const init = async () => {
    await SettingsService.GetSettings().catch(() => false);
    fetchData();
  };

  /**
   * Fetch data from server: executionsList & update currentExecution
   */
  const fetchData = async () => {
    clearTimeout(initTimeoutRequest);
    const executionsList = await ResourcesService.GetExecutions().catch(
      () => []
    );

    setExecutions(executionsList);
    setIsAppLoading(false);
    if (!executionsList.length) {
      // no data
      initTimeoutRequest = setTimeout(fetchData, 5000);
      return false;
    }

    let executionId = getHistory("executionId");

    const inListIndex = executionsList.findIndex(
      (row) => row.ID === executionId
    );

    if (inListIndex === -1) {
      executionId = executionsList[0].ID;
      setHistory({
        executionId,
      });
    }

    setCurrentExecution(executionId);
  };

  /**
   * Will help detect if we have resource in scanning mode
   * @param {array} ResourcesList Resource List fetched from server
   */
  const getScanningResource = (ResourcesList) => {
    const resource = Object.values(ResourcesList).find(
      (row) => row.Status === 0
    );
    if (resource) {
      return titleDirective(resource.ResourceName);
    }

    return false;
  };

  /**
   * Triggered every-time executionId/filters changes and re-load resources list
   * @param  {string} currentExecution Current Selected Execution
   * @param  {array} filters  Filters List
   */
  const onCurrentExecutionChanged = async (currentExecution, filters = []) => {
    clearTimeout(fetchTimeoutRequest);
    setIsResourceListLoading(true);
    await getResources(currentExecution, filters);
    setIsResourceListLoading(false);
  };

  /**
   * Will fetch resource list from server
   * @param  {string} currentExecution Current Selected Execution
   * @param  {array} filters  Filters List
   */
  const getResources = async (currentExecution, filters = []) => {
    const ResourcesList = await ResourcesService.Summary(
      currentExecution,
      filters
    ).catch(() => false);

    const scanningResource = getScanningResource(ResourcesList);
    if (scanningResource) {
      setIsScanning(true);
    } else {
      setIsScanning(false);
    }

    if (scanningResource) {
      fetchTimeoutRequest = setTimeout(
        () => getResources(currentExecution, filters),
        5000
      );
    }

    setResources(ResourcesList);
    return true;
  };

  /**
   * Initial Load - set baseURL using Settings Api
   */
  useEffect(() => {
    init();
  }, []);

  /**
   * currentExecution Change detection
   */
  useEffect(() => {
    if (currentExecution) {
      onCurrentExecutionChanged(currentExecution, filters);
    }
  }, [currentExecution]);

  /**
   * filters Change detection
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    // remove resource from filters detection
    const filtersList = JSON.stringify(
      filters.filter((row) => row.type !== "resource")
    );
    if (filtersList !== lastFiltersSearched) {
      lastFiltersSearched = filtersList;
      (async () =>
        await onCurrentExecutionChanged(currentExecution, filters))();
    }
  }, [filters]);

  return <Fragment></Fragment>;
};

DataFacotry.defaultProps = {};
DataFacotry.propTypes = {
  setExecutions: PropTypes.func,
  setIsAppLoading: PropTypes.func,
  setIsResourceListLoading: PropTypes.func,
  setIsScanning: PropTypes.func,
  setResources: PropTypes.func,
  setCurrentExecution: PropTypes.func,

  resources: PropTypes.object,
  filters: PropTypes.array,
  currentExecution: PropTypes.string,

  setScanning: PropTypes.func,

  isScanning: PropTypes.bool,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  currentExecution: state.executions.current,
  filters: state.filters.filters,
  isScanning: state.executions.isScanning,
});

const mapDispatchToProps = (dispatch) => ({
  setExecutions: (data) => dispatch({ type: "EXECUTION_LIST", data }),
  setIsAppLoading: (isLoading) =>
    dispatch({ type: "IS_APP_LOADING", isLoading }),
  setIsResourceListLoading: (isLoading) =>
    dispatch({ type: "IS_RESOURCE_LIST_LOADING", isLoading }),
  setIsScanning: (isScanning) => dispatch({ type: "IS_SCANNING", isScanning }),
  setResources: (data) => dispatch({ type: "RESOURCE_LIST", data }),
  setCurrentExecution: (id) => dispatch({ type: "EXECUTION_SELECTED", id }),
});

export default connect(mapStateToProps, mapDispatchToProps)(DataFacotry);
