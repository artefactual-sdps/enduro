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
 * @interface MonitorEventEvent
 */
export interface MonitorEventEvent {
    /**
     * Union type name, one of:
     * - "monitor_ping_event"
     * - "package_created_event"
     * - "package_updated_event"
     * - "package_status_updated_event"
     * - "package_location_updated_event"
     * - "preservation_action_created_event"
     * - "preservation_action_updated_event"
     * - "preservation_task_created_event"
     * - "preservation_task_updated_event"
     * @type {string}
     * @memberof MonitorEventEvent
     */
    type: MonitorEventEventTypeEnum;
    /**
     * JSON encoded union value
     * @type {string}
     * @memberof MonitorEventEvent
     */
    value: string;
}


/**
 * @export
 */
export const MonitorEventEventTypeEnum = {
    MonitorPingEvent: 'monitor_ping_event',
    PackageCreatedEvent: 'package_created_event',
    PackageUpdatedEvent: 'package_updated_event',
    PackageStatusUpdatedEvent: 'package_status_updated_event',
    PackageLocationUpdatedEvent: 'package_location_updated_event',
    PreservationActionCreatedEvent: 'preservation_action_created_event',
    PreservationActionUpdatedEvent: 'preservation_action_updated_event',
    PreservationTaskCreatedEvent: 'preservation_task_created_event',
    PreservationTaskUpdatedEvent: 'preservation_task_updated_event'
} as const;
export type MonitorEventEventTypeEnum = typeof MonitorEventEventTypeEnum[keyof typeof MonitorEventEventTypeEnum];


/**
 * Check if a given object implements the MonitorEventEvent interface.
 */
export function instanceOfMonitorEventEvent(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "type" in value;
    isInstance = isInstance && "value" in value;

    return isInstance;
}

export function MonitorEventEventFromJSON(json: any): MonitorEventEvent {
    return MonitorEventEventFromJSONTyped(json, false);
}

export function MonitorEventEventFromJSONTyped(json: any, ignoreDiscriminator: boolean): MonitorEventEvent {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'type': json['Type'],
        'value': json['Value'],
    };
}

export function MonitorEventEventToJSON(value?: MonitorEventEvent | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'Type': value.type,
        'Value': value.value,
    };
}

