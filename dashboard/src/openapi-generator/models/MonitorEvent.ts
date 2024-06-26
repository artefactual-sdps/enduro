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
import type { MonitorEventEvent } from './MonitorEventEvent';
import {
    MonitorEventEventFromJSON,
    MonitorEventEventFromJSONTyped,
    MonitorEventEventToJSON,
} from './MonitorEventEvent';

/**
 * 
 * @export
 * @interface MonitorEvent
 */
export interface MonitorEvent {
    /**
     * 
     * @type {MonitorEventEvent}
     * @memberof MonitorEvent
     */
    event?: MonitorEventEvent;
}

/**
 * Check if a given object implements the MonitorEvent interface.
 */
export function instanceOfMonitorEvent(value: object): boolean {
    let isInstance = true;

    return isInstance;
}

export function MonitorEventFromJSON(json: any): MonitorEvent {
    return MonitorEventFromJSONTyped(json, false);
}

export function MonitorEventFromJSONTyped(json: any, ignoreDiscriminator: boolean): MonitorEvent {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'event': !exists(json, 'event') ? undefined : MonitorEventEventFromJSON(json['event']),
    };
}

export function MonitorEventToJSON(value?: MonitorEvent | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'event': MonitorEventEventToJSON(value.event),
    };
}

