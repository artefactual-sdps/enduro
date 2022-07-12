/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
import {
    EnduroPackagePreservationTaskResponseBody,
    EnduroPackagePreservationTaskResponseBodyFromJSON,
    EnduroPackagePreservationTaskResponseBodyFromJSONTyped,
    EnduroPackagePreservationTaskResponseBodyToJSON,
} from './EnduroPackagePreservationTaskResponseBody';

/**
 * PreservationAction describes a preservation action. (default view)
 * @export
 * @interface EnduroPackagePreservationActionResponseBody
 */
export interface EnduroPackagePreservationActionResponseBody {
    /**
     * 
     * @type {Date}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    completedAt?: Date;
    /**
     * 
     * @type {number}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    id: number;
    /**
     * 
     * @type {string}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    name: string;
    /**
     * 
     * @type {Date}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    startedAt: Date;
    /**
     * 
     * @type {string}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    status: EnduroPackagePreservationActionResponseBodyStatusEnum;
    /**
     * EnduroPackage-Preservation-TaskCollectionResponseBody is the result type for an array of EnduroPackage-Preservation-TaskResponseBody (default view)
     * @type {Array<EnduroPackagePreservationTaskResponseBody>}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    tasks?: Array<EnduroPackagePreservationTaskResponseBody>;
    /**
     * 
     * @type {string}
     * @memberof EnduroPackagePreservationActionResponseBody
     */
    workflowId: string;
}


/**
 * @export
 */
export const EnduroPackagePreservationActionResponseBodyStatusEnum = {
    Unspecified: 'unspecified',
    Complete: 'complete',
    Processing: 'processing',
    Failed: 'failed'
} as const;
export type EnduroPackagePreservationActionResponseBodyStatusEnum = typeof EnduroPackagePreservationActionResponseBodyStatusEnum[keyof typeof EnduroPackagePreservationActionResponseBodyStatusEnum];


export function EnduroPackagePreservationActionResponseBodyFromJSON(json: any): EnduroPackagePreservationActionResponseBody {
    return EnduroPackagePreservationActionResponseBodyFromJSONTyped(json, false);
}

export function EnduroPackagePreservationActionResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroPackagePreservationActionResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'completedAt': !exists(json, 'completed_at') ? undefined : (new Date(json['completed_at'])),
        'id': json['id'],
        'name': json['name'],
        'startedAt': (new Date(json['started_at'])),
        'status': json['status'],
        'tasks': !exists(json, 'tasks') ? undefined : ((json['tasks'] as Array<any>).map(EnduroPackagePreservationTaskResponseBodyFromJSON)),
        'workflowId': json['workflow_id'],
    };
}

export function EnduroPackagePreservationActionResponseBodyToJSON(value?: EnduroPackagePreservationActionResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'completed_at': value.completedAt === undefined ? undefined : (value.completedAt.toISOString()),
        'id': value.id,
        'name': value.name,
        'started_at': (value.startedAt.toISOString()),
        'status': value.status,
        'tasks': value.tasks === undefined ? undefined : ((value.tasks as Array<any>).map(EnduroPackagePreservationTaskResponseBodyToJSON)),
        'workflow_id': value.workflowId,
    };
}

