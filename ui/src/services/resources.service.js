import { http } from './request.service'

export const ResourcesService = {
    GetExecutions,
    Summary,
    GetContent,
};

/**
 * Get executions data
 */
function GetExecutions() {
    return http.send(`api/v1/executions`, `get`).then(this.handleResponse).then(response => {
        return response;
    })
}

/**
 * Get resources metadata
 */
function Summary(executionID) {
    return http.send(`api/v1/summary?filter_ExecutionID=${executionID}`, `get`).then(this.handleResponse).then(response => {
        return response;
    })
}

/**
 * Get resource data
 * @param {string} name
 */
function GetContent(name, executionID) {

    return http.send(`api/v1/resources/${name}?executionID=${executionID}`, `get`).then(this.handleResponse).then(response => {
        return response;
    })

}