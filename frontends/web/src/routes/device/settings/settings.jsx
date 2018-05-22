import { Component } from 'preact';
import { ButtonLink } from '../../../components/forms';
import { translate } from 'react-i18next';
import { apiGet } from '../../../utils/request';
import Reset from './components/reset';
import MobilePairing from './components/mobile-pairing';
import DeviceLock from './components/device-lock';
import UpgradeFirmware from './components/upgradefirmware';
import Footer from '../../../components/footer/footer';

@translate()
export default class Settings extends Component {
    state = {
        firmwareVersion: null,
        lock: true,
    }

    componentDidMount() {
        apiGet('devices/' + this.props.deviceID + '/info').then(({ version, sdcard, lock }) => {
            this.setState({
                firmwareVersion: version.replace('v', ''),
                locked: lock,
            });
            // if (sdcard) alert('Keep the SD card stored securely unless you want to manage backups.');
        });
    }

    render({
        t,
        deviceID,
    }, {
        firmwareVersion,
        lock,
    }) {
        return (
            <div class="container">
                <div class="headerContainer">
                    <div class="header">
                        <h2>{t('deviceSettings.title')}</h2>
                    </div>
                </div>
                <div class="innerContainer">
                    <div class="content flex flex-column flex-start">
                        <div class={['flex', 'flex-row', 'flex-between', 'flex-1'].join(' ')}>
                            <ButtonLink primary href={`/manage-backups/${deviceID}`} disabled={lock}>{t('device.manageBackups')}</ButtonLink>
                            <MobilePairing deviceID={deviceID} disabled={lock}/>
                            <DeviceLock deviceID={deviceID} />
                            <UpgradeFirmware deviceID={deviceID} currentVersion={firmwareVersion} />
                            <Reset deviceID={deviceID} />
                        </div>
                        <Footer>
                            { firmwareVersion && <p>Firmware Version: {firmwareVersion}</p>}
                        </Footer>
                    </div>
                </div>
            </div>
        );
    }
}
