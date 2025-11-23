// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import PropTypes from 'prop-types';
import React, { useState } from 'react';
import { SketchPicker } from 'react-color';
import '../../styles/main.css';

const ColorMapEditor = ({ label, description, value, onChange, colorOptions }) => {
    const [colors, setColors] = useState(value || {});
    const [showPicker, setShowPicker] = useState({});

    const handleColorChange = (key, color) => {
        const newColors = {
            ...colors,
            [key]: color.hex,
        };
        setColors(newColors);
        onChange(newColors);
    };

    const handleRemoveColor = (key) => {
        const newColors = { ...colors };
        delete newColors[key];
        setColors(newColors);
        onChange(newColors);
    };

    const togglePicker = (key) => {
        setShowPicker({
            ...showPicker,
            [key]: !showPicker[key],
        });
    };

    return (
        <div className='color-map-editor'>
            <label className='control-label margin-bottom x2'>
                {label}
            </label>
            {description && (
                <div className='help-text'>
                    {description}
                </div>
            )}
            <div className='color-map-grid'>
                {colorOptions.map((option) => {
                    const currentColor = colors[option.key] || option.default;
                    return (
                        <div key={option.key} className='color-map-row'>
                            <div className='color-map-label'>
                                <strong>{option.label}</strong>
                                <span className='help-text'>{option.description}</span>
                            </div>
                            <div className='color-map-controls'>
                                <div className='color-swatch-container'>
                                    <button
                                        type='button'
                                        className='color-swatch'
                                        style={{ backgroundColor: currentColor }}
                                        onClick={() => togglePicker(option.key)}
                                        title={currentColor}
                                    >
                                        <span className='color-hex'>{currentColor}</span>
                                    </button>
                                    {showPicker[option.key] && (
                                        <div className='color-picker-popover'>
                                            <div
                                                className='color-picker-cover'
                                                onClick={() => togglePicker(option.key)}
                                            />
                                            <SketchPicker
                                                color={currentColor}
                                                onChange={(color) => handleColorChange(option.key, color)}
                                                disableAlpha={true}
                                            />
                                        </div>
                                    )}
                                </div>
                                {colors[option.key] && colors[option.key] !== option.default && (
                                    <button
                                        type='button'
                                        className='btn btn-sm btn-default'
                                        onClick={() => handleRemoveColor(option.key)}
                                        title='Reset to default'
                                    >
                                        ↺ Reset
                                    </button>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

ColorMapEditor.propTypes = {
    label: PropTypes.string.isRequired,
    description: PropTypes.string,
    value: PropTypes.object,
    onChange: PropTypes.func.isRequired,
    colorOptions: PropTypes.arrayOf(
        PropTypes.shape({
            key: PropTypes.string.isRequired,
            label: PropTypes.string.isRequired,
            description: PropTypes.string,
            default: PropTypes.string.isRequired,
        })
    ).isRequired,
};

export default ColorMapEditor;
