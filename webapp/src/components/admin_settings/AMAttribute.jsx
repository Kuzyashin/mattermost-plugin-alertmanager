import React, { useState, useEffect} from 'react';
import PropTypes from 'prop-types';
import crypto from 'crypto';

const AMAttribute = (props) => {
    const initialSettings = props.attributes === undefined || Object.keys(props.attributes).length === 0 ? {
        alertmanagerurl: "",
        channel: "",
        team: "",
        token: "",
        enableactions: false,
        severitymentions: "",
        firingtemplate: "",
        resolvedtemplate: "",
    } : {
        alertmanagerurl: props.attributes.alertmanagerurl? props.attributes.alertmanagerurl: "",
        channel: props.attributes.channel? props.attributes.channel : "",
        team: props.attributes.team ? props.attributes.team: "",
        token: props.attributes.token? props.attributes.token: "",
        enableactions: props.attributes.enableactions? props.attributes.enableactions: false,
        severitymentions: props.attributes.severitymentions? JSON.stringify(props.attributes.severitymentions): "",
        firingtemplate: props.attributes.firingtemplate? props.attributes.firingtemplate: "",
        resolvedtemplate: props.attributes.resolvedtemplate? props.attributes.resolvedtemplate: "",
    };

    const initErrors = {
        teamError: false,
        channelError: false,
        urlError: false
    };

    const [ settings, setSettings ] = useState(initialSettings);
    const [ hasError, setHasError ] = useState(initErrors);

    const handleTeamNameInput = (e) => {
        let newSettings = {...settings};

        if (!e.target.value || e.target.value.trim() === '') {
            setHasError({...hasError, teamError: true});
        } else {
            setHasError({...hasError, teamError: false});
        }

        newSettings = {...newSettings, team: e.target.value};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleChannelNameInput = (e) => {
        let newSettings = {...settings};

        if (!e.target.value || e.target.value.trim() === '') {
            setHasError({...hasError, channelError: true});
        } else {
            setHasError({...hasError, channelError: false});
        }

        newSettings = {...newSettings, channel: e.target.value};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleURLInput = (e) => {
        let newSettings = {...settings};

        if (!e.target.value || e.target.value.trim() === '') {
            setHasError({...hasError, urlError: true});
        } else {
            setHasError({...hasError, urlError: false});
        }

        newSettings = {...newSettings, alertmanagerurl: e.target.value};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleEnableActionsChange = (e) => {
        let newSettings = {...settings};
        newSettings = {...newSettings, enableactions: e.target.checked};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleSeverityMentionsInput = (e) => {
        let newSettings = {...settings};
        let severityMentionsValue = e.target.value;

        // Try to parse JSON for validation
        try {
            if (severityMentionsValue.trim() !== '') {
                JSON.parse(severityMentionsValue);
            }
        } catch (err) {
            // Invalid JSON, but still save the string for user to fix
        }

        newSettings = {...newSettings, severitymentions: severityMentionsValue};

        setSettings(newSettings);

        // Convert string to object when saving
        let attributesToSave = {...newSettings};
        if (severityMentionsValue.trim() !== '') {
            try {
                attributesToSave.severitymentions = JSON.parse(severityMentionsValue);
            } catch (err) {
                // Keep as string if invalid JSON
            }
        } else {
            attributesToSave.severitymentions = {};
        }

        props.onChange({id: props.id, attributes: attributesToSave});
    }

    const handleFiringTemplateInput = (e) => {
        let newSettings = {...settings};
        newSettings = {...newSettings, firingtemplate: e.target.value};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleResolvedTemplateInput = (e) => {
        let newSettings = {...settings};
        newSettings = {...newSettings, resolvedtemplate: e.target.value};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes: newSettings});
    }

    const handleDelete = (e) => {
        props.onDelete(props.id);
    }

    const regenerateToken = (e) => {
        e.preventDefault();

        // Generate a 32 byte tokes. It must not include '*' and '/'.
        // Copied from https://github.com/mattermost/mattermost-webapp/blob/33661c60bd05d708bcf85a49dad4d9fb3a39a75b/components/admin_console/generated_setting.tsx#L41
        const token = crypto.randomBytes(256).toString('base64').substring(0, 32).replaceAll('+', '-').replaceAll('/', '_');

        let newSettings = {...settings};
        newSettings = {...newSettings, token: token};

        setSettings(newSettings);
        props.onChange({id: props.id, attributes:newSettings});
    }

    const generateSimpleStringInputSetting = ( title, settingName, onChangeFunction, helpTextJSX) => {
        return (
            <div className="form-group" >
            <label className="control-label col-sm-4">
                {title}
            </label>
            <div className="col-sm-8">
                <input
                    id={`PluginSettings.Plugins.alertmanager.${settingName + "." + settings.id}`}
                    className="form-control"
                    type="input"
                    onChange={onChangeFunction}
                    value={settings[settingName]}
                />
                <div className="help-text">
                    {helpTextJSX}
                </div>
            </div>
        </div>
        );
    }

    const generateCheckboxSetting = ( title, settingName, onChangeFunction, helpTextJSX) => {
        return (
            <div className="form-group" >
            <label className="control-label col-sm-4">
                {title}
            </label>
            <div className="col-sm-8">
                <input
                    id={`PluginSettings.Plugins.alertmanager.${settingName + "." + settings.id}`}
                    type="checkbox"
                    onChange={onChangeFunction}
                    checked={settings[settingName]}
                />
                <div className="help-text">
                    {helpTextJSX}
                </div>
            </div>
        </div>
        );
    }

    const generateTextareaSetting = ( title, settingName, onChangeFunction, helpTextJSX) => {
        return (
            <div className="form-group" >
            <label className="control-label col-sm-4">
                {title}
            </label>
            <div className="col-sm-8">
                <textarea
                    id={`PluginSettings.Plugins.alertmanager.${settingName + "." + settings.id}`}
                    className="form-control"
                    rows="3"
                    onChange={onChangeFunction}
                    value={settings[settingName]}
                />
                <div className="help-text">
                    {helpTextJSX}
                </div>
            </div>
        </div>
        );
    }

    const generateGeneratedFieldSetting = ( title, settingName, regenerateFunction, regenerateText, helpTextJSX) => {
        return (<div className="form-group" >
        <label className="control-label col-sm-4">
            {title}
        </label>
        <div className="col-sm-8">
            <div
                id={`PluginSettings.Plugins.alertmanager.${settingName + "." + settings.id}`}
                className="form-control disabled"
                >
                {settings[settingName] !== undefined && settings[settingName] !== ""? settings[settingName] : <span className="placeholder-text"></span>}
            </div>
            <div className="help-text">
                {helpTextJSX}
            </div>
            <div className="help-text">
                <button
                    type="button"
                    className="btn btn-default"
                    onClick={regenerateFunction}
                >
                    <span>{regenerateText}</span>
                </button>
            </div>
        </div>
    </div>);
    }

    const hasAnyError = () => {
        return Object.values(hasError).findIndex(item => item) !== -1;
    }

    return (
        <div id={`setting_${props.id}`} className={`alert-setting ${hasAnyError() ? 'alert-setting--with-error' : ''}`}>
            <div className='alert-setting__controls'>
                <div className='alert-setting__order-number'>{`#${props.id}`}</div>
                <div id={`delete_${props.id}`} className='delete-setting btn btn-default' onClick={handleDelete}>{` X `}</div>
            </div>
            { hasAnyError() && <div className='alert-setting__error-text'>{`Attribute cannot be empty.`}</div> }
            <div className='alert-setting__content'>
                <div>
                    { generateSimpleStringInputSetting(
                        "Team Name:",
                        "team",
                        handleTeamNameInput,
                        (<span>{"Team you want to send messages to. Use the team name such as \'my-team\', instead of the display name."}</span>)
                        )
                    }

                    { generateSimpleStringInputSetting(
                        "Channel Name:",
                        "channel",
                        handleChannelNameInput,
                        (<span>{"Channel you want to send messages to. Use the channel name such as 'town-square', instead of the display name. If you specify a channel that does not exist, this plugin creates a new channel with that name."}</span>)
                        )
                    }

                    { generateGeneratedFieldSetting(
                        "Token:",
                        "token",
                        regenerateToken,
                        "Regenerate",
                        (<span>{"The token used to configure the webhook for AlertManager. The token is validates for each webhook request by the Mattermost server."}</span>)
                        )
                    }

                    { generateSimpleStringInputSetting(
                        "AlertManager URL:",
                        "alertmanagerurl",
                        handleURLInput,
                        (<span>{"The URL of your AlertManager instance, e.g. \'"}<a href="http://alertmanager.example.com/" rel="noopener noreferrer" target="_blank">{"http://alertmanager.example.com/"}</a>{"\'"}</span>)
                        )
                    }

                    { generateCheckboxSetting(
                        "Enable Action Buttons:",
                        "enableactions",
                        handleEnableActionsChange,
                        (<span>{"Enable interactive action buttons (Silence/ACK/UNACK) on alert posts"}</span>)
                        )
                    }

                    { generateTextareaSetting(
                        "Severity Mentions:",
                        "severitymentions",
                        handleSeverityMentionsInput,
                        (<span>{"JSON object mapping severity levels to mentions, e.g. "}<code>{'{"critical": "@devops-oncall", "warning": "@devops"}'}</code></span>)
                        )
                    }

                    { generateTextareaSetting(
                        "Firing Alert Template:",
                        "firingtemplate",
                        handleFiringTemplateInput,
                        (<span>{"Custom Go template for firing alerts. Leave empty to use default formatting. Available fields: .Labels, .Annotations, .StartsAt, .GeneratorURL"}</span>)
                        )
                    }

                    { generateTextareaSetting(
                        "Resolved Alert Template:",
                        "resolvedtemplate",
                        handleResolvedTemplateInput,
                        (<span>{"Custom Go template for resolved alerts. Leave empty to use default formatting. Available fields: .Labels, .Annotations, .StartsAt, .EndsAt"}</span>)
                        )
                    }
                </div>
            </div>
        </div>
    );
}

AMAttribute.propTypes = {
    id: PropTypes.string.isRequired,
    orderNumber: PropTypes.number.isRequired,
    attributes: PropTypes.object,
    onChange: PropTypes.func.isRequired
}

export default AMAttribute;
