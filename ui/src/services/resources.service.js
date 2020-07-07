import { http } from './request.service'

export const ResourcesService = {
    GetExecutions,
    Summary,
    GetContent,
};



const getTransformedFilters = (filters) => {

    const params = {};
    filters.forEach(filter => {
        if (filter.id.substr(0,8) === 'resource') {
            return; //skip
        }
        const [key, value] = filter.id.split(' : ');
        params[`filter_Data.Tag.${key.toLowerCase()}`] = value.toLowerCase();
    });
    return params;
}


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
function Summary(executionID, filters = []) {
    const params = {
        ...{ executionID } , 
        ...getTransformedFilters(filters) 
    };
    const searchParams = new window.URLSearchParams(params).toString();

    return http.send(`api/v1/summary/${executionID}?${searchParams}`, `get`).then(this.handleResponse).then(response => {
        return response;
    })
}

/**
 * Get resource data
 * @param {string} name
 */
function GetContent(name, executionID, filters = []) {

    const params = {
        ...{ executionID } , 
        ...getTransformedFilters(filters) 
    };
    const searchParams = new window.URLSearchParams(params).toString();

    return http.send(`api/v1/resources/${name}?${searchParams}`, `get`).then(this.handleResponse).then(response => {
        return response;
    })

}