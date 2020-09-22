import React, { Fragment, useEffect } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { ResourcesService } from "services/resources.service";
import { SettingsService } from "services/settings.service";
import { titleDirective } from "utils/Title";
import { getHistory, setHistory } from "../utils/History";

let fetchTimeoutRequest = false;
let fetchTableTimeoutRequest = false;
let initTimeoutRequest = false;
let lastFiltersSearched = "[]";

/**
 * will show a scanning message if some of the resources are still in progress
 * {
 * @param  {func} setExecutions Update Executions list
 * @param  {string} currentExecution Current Selected Execution
 * @param  {func} setCurrentExecution Update Current Execution
 *
 * @param  {string} currentResource Current selected resource
 * @param  {func} setResources Update Resources List
 * @param  {func} setCurrentResourceData Update current resource data
 * @param  {func} setIsResourceListLoading  update isLoading state for resources
 * @param  {func} setIsResourceTableLoading  update isLoading state for resources table
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

  currentResource,
  setResources,
  setCurrentResourceData,
  setIsResourceListLoading,
  setIsResourceTableLoading,

  filters,
  resources,
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

    if (currentResource) {
      await onCurrentResourceChanged(
        currentResource,
        currentExecution,
        filters
      );
    }
  };
  /**
   * Triggered every-time currentResource changes and re-load resources table data
   * @param  {string} currentExecution Current Selected Execution
   * @param  {array} filters  Filters List
   */
  const onCurrentResourceChanged = async (
    currentResource,
    currentExecution,
    filters = []
  ) => {
    clearTimeout(fetchTableTimeoutRequest);
    setIsResourceTableLoading(true);
    await getResourceTable(currentResource, currentExecution, filters);
    setIsResourceTableLoading(false);
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
   * Will fetch resource data from server
   * @param  {string} currentResource Current Selected Resource
   * @param  {string} currentExecution Current Selected Execution
   * @param  {array} filters  Filters List
   */
  const getResourceTable = async (
    currentResource,
    currentExecution,
    filters = []
  ) => {
    clearTimeout(fetchTableTimeoutRequest);
    const ResourceRows = await ResourcesService.GetContent(
      currentResource,
      currentExecution,
      filters
    ).catch(() => []);

    let rows = [];
    if (ResourceRows && ResourceRows.length) {
      rows = ResourceRows.map((row) => row.Data);
    }
    setCurrentResourceData(rows);

    const resourceInfo = resources[currentResource];
    // resource in scanning mode - refresh
    if (resourceInfo && resourceInfo.Status == 0) {
      fetchTableTimeoutRequest = setTimeout(
        () => getResourceTable(currentResource, currentExecution, filters),
        5000
      );
    } else {
      clearTimeout(fetchTableTimeoutRequest);
    }

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
   * currentResource Change detection
   */
  useEffect(() => {
    if (currentExecution && currentResource) {
      onCurrentResourceChanged(currentResource, currentExecution, filters);
    }
  }, [currentResource]);

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
  setIsResourceTableLoading: PropTypes.func,
  setIsScanning: PropTypes.func,
  setResources: PropTypes.func,
  setCurrentResourceData: PropTypes.func,
  setCurrentExecution: PropTypes.func,

  currentResource: PropTypes.string,
  resources: PropTypes.object,
  filters: PropTypes.array,
  currentExecution: PropTypes.string,

  setScanning: PropTypes.func,

  isScanning: PropTypes.bool,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  currentResource: state.resources.currentResource,
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
  setIsResourceTableLoading: (isLoading) =>
    dispatch({ type: "IS_RESOURCE_TABLE_LOADING", isLoading }),
  setIsScanning: (isScanning) => dispatch({ type: "IS_SCANNING", isScanning }),
  setResources: (data) => dispatch({ type: "RESOURCE_LIST", data }),
  setCurrentExecution: (id) => dispatch({ type: "EXECUTION_SELECTED", id }),
  setCurrentResourceData: (data) =>
    dispatch({ type: "SET_CURRENT_RESOURCE_DATA", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(DataFacotry);
