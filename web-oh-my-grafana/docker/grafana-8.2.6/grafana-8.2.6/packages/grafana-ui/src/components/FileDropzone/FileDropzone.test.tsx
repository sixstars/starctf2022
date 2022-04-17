import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { FileDropzone } from './FileDropzone';
import { REMOVE_FILE } from './FileListItem';

const file = ({
  fileBits = JSON.stringify({ ping: true }),
  fileName = 'ping.json',
  options = { type: 'application/json' },
}) => new File([fileBits], fileName, options);

const files = [
  file({}),
  file({ fileName: 'pong.json' }),
  file({ fileBits: 'something', fileName: 'something.jpg', options: { type: 'image/jpeg' } }),
];

describe('The FileDropzone component', () => {
  afterEach(() => {
    jest.resetAllMocks();
  });

  it('should show the default text of the dropzone component when no props passed', () => {
    render(<FileDropzone />);

    expect(screen.getByText('Upload file')).toBeInTheDocument();
  });

  it('should show accepted file type when passed in the options as a string', () => {
    render(<FileDropzone options={{ accept: '.json' }} />);

    expect(screen.getByText('Accepted file type: .json')).toBeInTheDocument();
  });

  it('should show accepted file types when passed in the options as a string array', () => {
    render(<FileDropzone options={{ accept: ['.json', '.txt'] }} />);

    expect(screen.getByText('Accepted file types: .json, .txt')).toBeInTheDocument();
  });

  it('should handle file removal from the list', async () => {
    render(<FileDropzone />);

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData(files));

    expect(await screen.findAllByLabelText(REMOVE_FILE)).toHaveLength(3);

    fireEvent.click(screen.getAllByLabelText(REMOVE_FILE)[0]);

    expect(await screen.findAllByLabelText(REMOVE_FILE)).toHaveLength(2);
  });

  it('should overwrite selected file when multiple false', async () => {
    render(<FileDropzone options={{ multiple: false }} />);

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([file({})]));

    expect(await screen.findAllByLabelText(REMOVE_FILE)).toHaveLength(1);
    expect(screen.getByText('ping.json')).toBeInTheDocument();

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([file({ fileName: 'newFile.jpg' })]));

    expect(await screen.findByText('newFile.jpg')).toBeInTheDocument();
    expect(screen.getAllByLabelText(REMOVE_FILE)).toHaveLength(1);
  });

  it('should use the passed readAs prop with the FileReader API', async () => {
    render(<FileDropzone readAs="readAsDataURL" />);
    const fileReaderSpy = jest.spyOn(FileReader.prototype, 'readAsDataURL');

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([file({})]));

    expect(await screen.findByText('ping.json')).toBeInTheDocument();
    expect(fileReaderSpy).toBeCalled();
  });

  it('should use the readAsText FileReader API if no readAs prop passed', async () => {
    render(<FileDropzone />);
    const fileReaderSpy = jest.spyOn(FileReader.prototype, 'readAsText');

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([file({})]));

    expect(await screen.findByText('ping.json')).toBeInTheDocument();
    expect(fileReaderSpy).toBeCalled();
  });

  it('should use the onDrop that is passed', async () => {
    const onDrop = jest.fn();
    const fileToUpload = file({});
    render(<FileDropzone options={{ onDrop }} />);
    const fileReaderSpy = jest.spyOn(FileReader.prototype, 'readAsText');

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([fileToUpload]));

    expect(await screen.findByText('ping.json')).toBeInTheDocument();
    expect(fileReaderSpy).not.toBeCalled();
    expect(onDrop).toBeCalledWith([fileToUpload], [], expect.anything());
  });

  it('should show children inside the dropzone', () => {
    const component = (
      <FileDropzone>
        <p>Custom dropzone text</p>
      </FileDropzone>
    );
    render(component);

    screen.getByText('Custom dropzone text');
  });

  it('should handle file list overwrite when fileListRenderer is passed', async () => {
    render(<FileDropzone fileListRenderer={() => null} />);

    dispatchEvt(screen.getByTestId('dropzone'), 'drop', mockData([file({})]));

    // need to await this in order to have the drop finished
    await screen.findByTestId('dropzone');

    expect(screen.queryByText('ping.json')).not.toBeInTheDocument();
  });
});

function dispatchEvt(node: HTMLElement, type: string, data: any) {
  const event = new Event(type, { bubbles: true });
  Object.assign(event, data);
  fireEvent(node, event);
}

function mockData(files: File[]) {
  return {
    dataTransfer: {
      files,
      items: files.map((file) => ({
        kind: 'file',
        type: file.type,
        getAsFile: () => file,
      })),
      types: ['Files'],
    },
  };
}
