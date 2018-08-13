import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import MoreVert from "@material-ui/icons/MoreVert";
import RootRef from "@material-ui/core/RootRef";
import TableCell from "@material-ui/core/TableCell";

import MenuController from "/components/controller/MenuController";

import ResourceDetails from "/components/partials/ResourceDetails";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

import EventStatusDescriptor from "/components/partials/EventStatusDescriptor";
import NamespaceLink from "/components/util/NamespaceLink";
import CheckStatusIcon from "/components/CheckStatusIcon";

class EventListItem extends React.PureComponent {
  static propTypes = {
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickSilencePair: PropTypes.func.isRequired,
    onClickSilenceEntity: PropTypes.func.isRequired,
    onClickSilenceCheck: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    event: PropTypes.shape({
      entity: PropTypes.shape({
        name: PropTypes.string.isRequired,
      }).isRequired,
      check: PropTypes.shape({
        name: PropTypes.string.isRequired,
      }).isRequired,
      timestamp: PropTypes.string.isRequired,
    }).isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventsListItem_event on Event {
        id
        isSilenced
        deleted @client
        check {
          name
          isSilenced
          history(first: 1) {
            status
          }
          ...EventStatusDescriptor_check
        }
        entity {
          name
        }
        namespace {
          organization
          environment
        }
        ...EventStatusDescriptor_event
      }

      ${EventStatusDescriptor.fragments.check}
      ${EventStatusDescriptor.fragments.event}
    `,
  };

  handleClickCheckbox = () => {
    this.props.onChangeSelected(!this.props.selected);
  };

  renderMenu = ({ close, anchorEl }) => {
    const { event } = this.props;

    return (
      <Menu open onClose={close} anchorEl={anchorEl}>
        <MenuItem
          key={"silence-Entity"}
          onClick={() => {
            this.props.onClickSilenceEntity();
            close();
          }}
        >
          Silence Entity
        </MenuItem>
        <MenuItem
          key={"silence-Check"}
          onClick={() => {
            this.props.onClickSilenceCheck();
            close();
          }}
        >
          Silence Check
        </MenuItem>
        <MenuItem
          key={"silence-pair"}
          onClick={() => {
            this.props.onClickSilencePair();
            close();
          }}
        >
          Silence Both
        </MenuItem>
        {event.check.isSilenced && (
          <MenuItem
            onClick={() => {
              this.props.onClickClearSilences();
              close();
            }}
          >
            Unsilence
          </MenuItem>
        )}
        {event.check.status !== 0 && (
          <MenuItem
            onClick={() => {
              this.props.onClickResolve();
              close();
            }}
          >
            Resolve
          </MenuItem>
        )}
      </Menu>
    );
  };

  render() {
    const { selected, event } = this.props;
    const { entity, check, timestamp } = event;

    // Try to determine if the failing check just started failing and if so
    // highlight the row.
    const incidentStarted =
      check.status > 0 &&
      check.history.length > 0 &&
      check.history[0].status !== check.status &&
      new Date(new Date(timestamp).valueOf() + 2500) >= new Date();

    return (
      <TableSelectableRow selected={selected} highlight={incidentStarted}>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            checked={selected}
            onChange={this.handleClickCheckbox}
          />
        </TableCell>
        <TableOverflowCell>
          <ResourceDetails
            icon={
              event.check && (
                <CheckStatusIcon
                  statusCode={event.check.status}
                  silenced={event.isSilenced}
                />
              )
            }
            title={
              <NamespaceLink
                namespace={event.namespace}
                to={`/events/${entity.name}/${check.name}`}
              >
                <strong>
                  {entity.name} › {check.name}
                </strong>
              </NamespaceLink>
            }
            details={
              <EventStatusDescriptor event={event} check={event.check} />
            }
          />
        </TableOverflowCell>
        <TableCell padding="checkbox">
          <MenuController renderMenu={this.renderMenu}>
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <MoreVert />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default EventListItem;
