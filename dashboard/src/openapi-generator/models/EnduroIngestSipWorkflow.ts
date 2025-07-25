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
import type { EnduroIngestSipTask } from './EnduroIngestSipTask';
import {
    EnduroIngestSipTaskFromJSON,
    EnduroIngestSipTaskFromJSONTyped,
    EnduroIngestSipTaskToJSON,
} from './EnduroIngestSipTask';

/**
 * SIPWorkflow describes a workflow of a SIP.
 * @export
 * @interface EnduroIngestSipWorkflow
 */
export interface EnduroIngestSipWorkflow {
    /**
     * 
     * @type {Date}
     * @memberof EnduroIngestSipWorkflow
     */
    completedAt?: Date;
    /**
     * Identifier of related SIP
     * @type {string}
     * @memberof EnduroIngestSipWorkflow
     */
    sipUuid: string;
    /**
     * 
     * @type {Date}
     * @memberof EnduroIngestSipWorkflow
     */
    startedAt: Date;
    /**
     * 
     * @type {string}
     * @memberof EnduroIngestSipWorkflow
     */
    status: EnduroIngestSipWorkflowStatusEnum;
    /**
     * 
     * @type {Array<EnduroIngestSipTask>}
     * @memberof EnduroIngestSipWorkflow
     */
    tasks?: Array<EnduroIngestSipTask>;
    /**
     * 
     * @type {string}
     * @memberof EnduroIngestSipWorkflow
     */
    temporalId: string;
    /**
     * 
     * @type {string}
     * @memberof EnduroIngestSipWorkflow
     */
    type: EnduroIngestSipWorkflowTypeEnum;
    /**
     * Identifier of the workflow
     * @type {string}
     * @memberof EnduroIngestSipWorkflow
     */
    uuid: string;
}


/**
 * @export
 */
export const EnduroIngestSipWorkflowStatusEnum = {
    Unspecified: 'unspecified',
    InProgress: 'in progress',
    Done: 'done',
    Error: 'error',
    Queued: 'queued',
    Pending: 'pending',
    Failed: 'failed'
} as const;
export type EnduroIngestSipWorkflowStatusEnum = typeof EnduroIngestSipWorkflowStatusEnum[keyof typeof EnduroIngestSipWorkflowStatusEnum];

/**
 * @export
 */
export const EnduroIngestSipWorkflowTypeEnum = {
    CreateAip: 'create aip',
    CreateAndReviewAip: 'create and review aip'
} as const;
export type EnduroIngestSipWorkflowTypeEnum = typeof EnduroIngestSipWorkflowTypeEnum[keyof typeof EnduroIngestSipWorkflowTypeEnum];


/**
 * Check if a given object implements the EnduroIngestSipWorkflow interface.
 */
export function instanceOfEnduroIngestSipWorkflow(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "sipUuid" in value;
    isInstance = isInstance && "startedAt" in value;
    isInstance = isInstance && "status" in value;
    isInstance = isInstance && "temporalId" in value;
    isInstance = isInstance && "type" in value;
    isInstance = isInstance && "uuid" in value;

    return isInstance;
}

export function EnduroIngestSipWorkflowFromJSON(json: any): EnduroIngestSipWorkflow {
    return EnduroIngestSipWorkflowFromJSONTyped(json, false);
}

export function EnduroIngestSipWorkflowFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroIngestSipWorkflow {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'completedAt': !exists(json, 'completed_at') ? undefined : (new Date(json['completed_at'])),
        'sipUuid': json['sip_uuid'],
        'startedAt': (new Date(json['started_at'])),
        'status': json['status'],
        'tasks': !exists(json, 'tasks') ? undefined : ((json['tasks'] as Array<any>).map(EnduroIngestSipTaskFromJSON)),
        'temporalId': json['temporal_id'],
        'type': json['type'],
        'uuid': json['uuid'],
    };
}

export function EnduroIngestSipWorkflowToJSON(value?: EnduroIngestSipWorkflow | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'completed_at': value.completedAt === undefined ? undefined : (value.completedAt.toISOString()),
        'sip_uuid': value.sipUuid,
        'started_at': (value.startedAt.toISOString()),
        'status': value.status,
        'tasks': value.tasks === undefined ? undefined : ((value.tasks as Array<any>).map(EnduroIngestSipTaskToJSON)),
        'temporal_id': value.temporalId,
        'type': value.type,
        'uuid': value.uuid,
    };
}

