/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 0.0.1
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface EnduroPoststorage
 */
export interface EnduroPoststorage {
    /**
     * 
     * @type {string}
     * @memberof EnduroPoststorage
     */
    taskQueue: string;
    /**
     * 
     * @type {string}
     * @memberof EnduroPoststorage
     */
    workflowName: string;
}

/**
 * Check if a given object implements the EnduroPoststorage interface.
 */
export function instanceOfEnduroPoststorage(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "taskQueue" in value;
    isInstance = isInstance && "workflowName" in value;

    return isInstance;
}

export function EnduroPoststorageFromJSON(json: any): EnduroPoststorage {
    return EnduroPoststorageFromJSONTyped(json, false);
}

export function EnduroPoststorageFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroPoststorage {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'taskQueue': json['task_queue'],
        'workflowName': json['workflow_name'],
    };
}

export function EnduroPoststorageToJSON(value?: EnduroPoststorage | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'task_queue': value.taskQueue,
        'workflow_name': value.workflowName,
    };
}
