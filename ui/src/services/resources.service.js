import { http } from './request.service'

export const ResourcesService = {
    Summary,
    GetContent,
};

/**
 * Get resources metadata
 */
function Summary() {
    return http.send(`api/v1/summary`, `get`).then(this.handleResponse).then(response => {
        return response;
    })
}

/**
 * Get resource data
 * @param {string} name
 */
function GetContent(name) {

    return http.send(`api/v1/resources/${name}`, `get`).then(this.handleResponse).then(response => {
        return response;
    })

}