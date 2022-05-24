import * as React from 'react';
import { useRef, useState, useEffect } from 'react';
import { StyledEngineProvider, ThemeProvider, createTheme } from '@mui/material/styles';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import TextField from '@mui/material/TextField';
import Grid from '@mui/material/Grid';
import TextareaAutosize from '@mui/material/TextareaAutosize';
import './style.css';
import logo from './logo.txt';
import git_image from './git.png';
import commit_image from './commit.png';
import yaml from 'js-yaml';
 
type User = {
    user: string,
    lastSHA1: string,
    reviewcommits?: string[]
}

type Repo = {
    provider: string,
    url: string,
    branch: string,
    lastSHA1: string,
    users?: User[],
    reviewcommits?: string[]
}

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

const server = "http://localho.st:1323";
const refresh_interval = 5000;

function isValidGitUrl(string: string) {
    let url;
    
    try {
      url = new URL(string);
    } catch (_) {
      return false;  
    }
  
    let validHttpUrl = url.protocol === "http:" || url.protocol === "https:";
    let re = /^(([A-Za-z0-9]+@|http(|s)\:\/\/)|(http(|s)\:\/\/[A-Za-z0-9]+@))([A-Za-z0-9.]+(:\d+)?)(?::|\/)([\d\/\w.-]+?)(\.git){1}$/;
    let giturl = re.test(string);

    if(validHttpUrl || giturl) {
        return true;
    }
    return false;
}

function TabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;

    return (
        <div role="tabpanel" hidden={value !== index} id={`simple-tabpanel-${index}`} aria-labelledby={`simple-tab-${index}`} {...other}>
            {value === index && (
            <Box sx={{ p: 3 }}>
                {children}
            </Box>
            )}
        </div>
    );
}

function a11yProps(index: number) {
    return {
        id: `simple-tab-${index}`,
        'aria-controls': `simple-tabpanel-${index}`,
    };
}

function ServerLog() {
    const [loglines, setLogLines] = useState<string[]>([]);
    const [refreshInterval, setRefreshInterval] = useState(refresh_interval);

    const get_server_log = async () => {
        const data = await fetch(server + "/serverlog", {
            method: "GET"
        });
        const jsonData = await data.json();
        setLogLines(jsonData.reverse());
    };

    useEffect(() => {
        if (refreshInterval && refreshInterval > 0) {
          const interval = setInterval(get_server_log, refreshInterval);
          return () => clearInterval(interval);
        }
      }, [refreshInterval]);

    useEffect(() => {  
        setRefreshInterval(refresh_interval)  
        get_server_log();
    }, []);

    return (
        <>
            <h3>Server log</h3>
            <pre className="serverLog">
            {
                loglines.map((item, i) =>  { return <p key={i}>{item}</p> })
            }
            </pre>
        </>
    );
}

function Commits() {
    const [result, setRepoResult] = useState<Repo[]>([]);
    const [refreshInterval, setRefreshInterval] = useState(refresh_interval);

    const get_repo_diff = async () => {
        const data = await fetch(server + "/config", {
            method: "GET"
        });
        const repos_yaml = await data.text();
        const repos_yaml_struct = yaml.load(repos_yaml)
        setRepoResult(repos_yaml_struct.repos);        
    };

    useEffect(() => {
        if (refreshInterval && refreshInterval > 0) {
          const interval = setInterval(get_repo_diff, refreshInterval);
          return () => clearInterval(interval);
        }
      }, [refreshInterval]);

    useEffect(() => {  
        setRefreshInterval(refresh_interval)  
        get_repo_diff();
    }, []);

    const reviewedCommit = async(event: any): Promise<void> => {
        event.preventDefault();
        console.log(event.target.id);

        const review_commit = async () => {
            const data = await fetch(server + "/review/" + event.target.id, {
                method: "GET"
            });
            const repos_yaml = await data.text();
            const repos_yaml_struct = yaml.load(repos_yaml);
            setRepoResult(repos_yaml_struct.repos);
        };

        review_commit();
    }

    const RepoUserCommits = (props: any): JSX.Element => (
        <li key={Math.random()}>
            {props.user.reviewcommits?.map((user_commit: any, user_commit_index: any, user_reviewcommits: any) => {
                let [user_title, user_url] = user_commit.split(",#,");
                let [user_commit_date, user_commit_title, commit_user] = user_title.split("||");
                let user_commit_ref = props.index.toString() + "_" + props.user_index.toString() + "_" + user_commit_index.toString();

                return (
                    <div key={Math.random()}>
                        <span>
                            <img src={commit_image} alt="commit image" width="14" height="14" className="commit_image" />
                            <a target="_blank" title={user_commit_title} href={user_url} className="linkCommit">{user_commit_title}</a>
                            <br />
                            <span className="commitByLine">on <b className="commitByLineDate">{user_commit_date}</b> by <b className="commitByLineUser">{commit_user}</b></span>
                        </span>
                        <span className="review_span">
                            {user_reviewcommits.length - 1 === user_commit_index ? <Button variant="outlined" className='review_button' id={user_commit_ref} size="large" onClick={reviewedCommit}>Reviewed</Button> : <span className='review_skip'>...</span>}
                        </span>
                    </div>
                );
            })}
        </li>
    )

    const RepoUsers = (props: any) => {
        return (
            <ol key={Math.random()} className="user_commit_list">
            {
                props.users?.map((user: any, user_index: any, users: any) => {
                    return <RepoUserCommits key={Math.random()} index={props.index} user={user} user_index={user_index} users={users} />
                })
            }
            </ol>
        )
    }

    const RepoCommits = (props: any) => {
        return (
            <ol key={Math.random()} className="commit_list">
            {                                                
                props.item.reviewcommits?.map((commit: any, commit_index: any, reviewcommits: any) => {
                    let [title, url] = commit.split(",#,");
                    let [date, commit_title, user] = title.split("||");
                    let commit_ref = props.index.toString() + "_" + commit_index.toString();
                    return (
                        <li key={Math.random()}> 
                            <span>                                           
                                <img src={commit_image} alt="commit image" width="14" height="14" className="commit_image"/>
                                <a target="_blank" title={commit_title} href={url}  className="linkCommit">{commit_title}</a>
                                <br/>
                                <span className="commitByLine">on <b className="commitByLineDate">{date}</b> by <b className="commitByLineUser">{user}</b></span>
                            </span>
                            <span className="review_span">
                            {
                                reviewcommits.length - 1 === commit_index ? 
                                    <Button variant="outlined" className='review_button' id={commit_ref} size="large" onClick={reviewedCommit}>Reviewed</Button> 
                                : <span className='review_skip'>...</span>
                            }
                            </span>
                        </li>
                    )
                })
            }
            </ol> 
        )
    }

    return (
        <>
            <h3>Latest commits</h3>
            <ol key={Math.random()} className="repo_list">
            {
                result.map((item, index) => 
                    <li key={Math.random()}><img src={git_image} alt="git logo" width="16" height="16" className="git_image"/>
                        <a target="_blank" title={item.url} href={item.url} className="linkRepo">{item.url}</a>
                        {
                            item.reviewcommits !== null ? <RepoCommits key={Math.random()} index={index} item={item} /> : <RepoUsers key={Math.random()} index={index} users={item.users} />
                        }
                    </li>
                )                            
            }
            </ol>
        </>
    );
}

function Config() {
    const [result, setResult] = useState<string>();
    const [pretty, setPrettyResult] = useState<string>();
    const [loading, setloadingResult] = useState<boolean>();
    const formRef = useRef()

    const [provider, setProvider] = useState<string>();
    const [url, setUrl] = useState<string>();
    const [branch, setBranch] = useState<string>();
    const [sha1, setSha1] = useState<string>();

    const [editProvider, setEditProvider] = useState<string>();
    const [editUrl, setEditUrl] = useState<string>();
    const [editBranch, setEditBranch] = useState<string>();
    const [editSha1, setEditSha1] = useState<string>();
    const [shrink, setShrink] = useState<boolean>();

    const get_config = async () => {
        const data = await fetch(server + "/pretty", {
            method: "GET"
        });
        const text = await data.text();
        const repos_yaml_struct = yaml.load(text);
        setPrettyResult(text);
        setResult(repos_yaml_struct.repos);     
    };

    useEffect(() => {
        setProvider("github");
        setUrl("");
        setBranch("main");
        setSha1("");
        setEditSha1("");
        setEditUrl("");
        setEditProvider("gitlab")
        get_config();
    }, []);

    const saveConfig = async(event: any) => {   
        setloadingResult(true);
        event.preventDefault();
        const data = new FormData(formRef.current)
        
        const yaml_fetcher = await fetch(server + "/update", {
            method: "POST",
            mode: 'cors', 
            body: data
        });
        
        const text = await yaml_fetcher.text();
        const repos_yaml_struct = yaml.load(text);
        setPrettyResult(text);
        setResult(repos_yaml_struct.repos);     
        setloadingResult(false);
    }

    const handleAddRepo = async(event: any) => {
        setloadingResult(true);
        event.preventDefault();

        const formData = new FormData();
        formData.append('provider', (provider == undefined ? "" : provider));
        formData.append('branch', (branch == undefined ? "" : branch));
        formData.append('sha1', (sha1 == undefined ? "" : sha1));
        formData.append('url', (url == undefined ? "" : url));
        
        const yaml_fetcher = await fetch(server + "/add", {
            method: "POST",
            mode: 'cors', 
            body: formData
        });
        
        await yaml_fetcher.text();
        get_config();  
        setloadingResult(false); 
    }

    const handleDeleteRepo = async(event: any) => {
        setloadingResult(true);
        event.preventDefault();

        const formData = new FormData();
        formData.append('url', (editUrl == undefined ? "" : editUrl));
        
        const yaml_fetcher = await fetch(server + "/delete", {
            method: "DELETE",
            mode: 'cors', 
            body: formData
        });
        
        await yaml_fetcher.text();
        get_config();
        setloadingResult(false); 
    }

    const handleEditRepo = async(event: any) => {
        setloadingResult(true);
        event.preventDefault();

        const formData = new FormData();
        formData.append('provider', (editProvider == undefined ? "" : editProvider));
        formData.append('branch', (editBranch == undefined ? "" : editBranch));
        formData.append('sha1', (editSha1 == undefined ? "" : editSha1));
        formData.append('url', (editUrl == undefined ? "" : editUrl));

        const yaml_fetcher = await fetch(server + "/edit", {
            method: "POST",
            mode: 'cors', 
            body: formData
        });
        
        await yaml_fetcher.text();
        get_config(); 
        setloadingResult(false); 
    }

    const editingSelectedChanged = async(event: any) => {
        event.preventDefault();
        const selected = event.target.value;
        const obj = result?.find(obj => { return obj.url === selected });

        setShrink(true);
        setEditProvider(obj.provider);
        setEditUrl(obj.url);
        setEditBranch(obj.branch);
        setEditSha1(obj.lastSHA1);

        console.log(obj.provider);
        console.log(obj.url);
        console.log(obj.branch);
        console.log(obj.lastSHA1);
    }

    return (
        <>
            <h3>Simple Editor</h3>

            <Grid container direction={"row"}>
                <Grid item xs={5} className="editorWrapper">
                    <h4>Add</h4>
                    <form onSubmit={handleAddRepo}> 
                        <FormControl fullWidth>
                            <InputLabel id="demo-simple-select-label">Select provider</InputLabel>
                            <Select 
                                labelId="demo-simple-select-label" 
                                id="demo-simple-select" 
                                label="Select provider"
                                defaultValue=""
                            >
                                <MenuItem value="github">github</MenuItem>
                                <MenuItem value="gitlab">gitlab</MenuItem>
                            </Select>
                        </FormControl>
                        <br /><br />
                        <Grid container direction={"column"}>
                            <Grid item>
                                <FormControl fullWidth variant="filled" margin="none">
                                    <TextField 
                                        id="outlined-basic" 
                                        label="Url" 
                                        variant="outlined" 
                                        helperText="The url to the repository, if you provide git://, please also setup ssh keys" 
                                        margin="none" 
                                        onInput={e=>setUrl(e.target.value)}
                                        error={isValidGitUrl(url) === false}
                                        defaultValue={url}
                                    /> 
                                </FormControl>
                            </Grid>
                            <Grid item>
                                <TextField 
                                    id="outlined-basic" 
                                    label="Branch" 
                                    variant="outlined" 
                                    helperText="The branch such as master or main" 
                                    margin="dense" 
                                    error={branch?.length < 2}
                                    onInput={e=>setBranch(e.target.value)}
                                    defaultValue={branch}
                                />
                            </Grid>
                            <Grid item>
                                <FormControl fullWidth variant="filled" margin="none">
                                    <TextField 
                                        id="outlined-basic" 
                                        label="Last Sha1" 
                                        variant="outlined" 
                                        helperText="The last SHA1 commit id reviewed" 
                                        margin="dense"  
                                        onInput={e=>setSha1(e.target.value)}
                                        error={sha1?.length != 40 && sha1?.length != 0}
                                        defaultValue={sha1}
                                    />
                                </FormControl>
                            </Grid>
                        </Grid> 
                        <br />
                        <Button variant="outlined" size="large" type="submit">
                            Add
                        </Button>
                    </form>  
                </Grid>                 
                <Grid item xs={1}>
                    &nbsp;
                </Grid>
                <Grid item xs={6} className="editorWrapper">
                    <h4>Edit</h4>
                    <form ref={formRef} onSubmit={saveConfig}>                 
                        <Grid container direction={"row"} spacing={1}>
                            <Grid item xs={10}>
                                <FormControl fullWidth>
                                    <InputLabel id="demo-simple-select-label">Select Repository</InputLabel>
                                    <Select labelId="demo-simple-select-label" 
                                        id="demo-simple-select" 
                                        label="Select Repository" 
                                        onChange={editingSelectedChanged} 
                                        onSelect={() => setShrink(true)}
                                        defaultValue="">
                                        {
                                            result?.map((item, index) => <MenuItem key={index} value={item.url}>{item.url}</MenuItem> )
                                        }
                                    </Select>
                                </FormControl>
                            </Grid>
                            <Grid item xs={2}>
                                <Button 
                                    variant="outlined" 
                                    color="error" 
                                    size="large" 
                                    style={{marginTop: '5px', marginRight: '0px',marginLeft: 'auto', display: "flex"}}
                                    disabled={editUrl == ""}
                                    onClick={handleDeleteRepo}
                                >
                                    Delete
                                </Button>
                            </Grid>
                        </Grid> 
                        <br />
                        <FormControl fullWidth>
                            <InputLabel id="demo-simple-select-label">Select provider</InputLabel>
                            <Select 
                                labelId="demo-simple-select-label" 
                                id="demo-simple-select" 
                                label="Select provider" 
                                onChange={e=>setEditBranch(e.target.value)}
                                value={editProvider ? editProvider : "github"}
                                disabled={editUrl == ""}>
                                <MenuItem value="github">github</MenuItem>
                                <MenuItem value="gitlab">gitlab</MenuItem>
                            </Select>
                        </FormControl>
                        <br /><br />
                        <Grid container direction={"column"} spacing={1}>
                            <Grid item>
                                <TextField id="outlined-basic" 
                                    label="Branch" 
                                    helperText="The branch such as master or main" 
                                    variant="outlined" 
                                    margin="dense" 
                                    value={editBranch ? editBranch : ""}
                                    error={editBranch?.length < 2}
                                    onInput={e=>setEditBranch(e.target.value)}
                                    disabled={editUrl == ""}
                                    InputLabelProps={{ shrink: shrink }}/>
                            </Grid>
                            <Grid item>
                                <FormControl fullWidth variant="filled" margin="none">
                                    <TextField 
                                        id="outlined-basic" 
                                        label="Last Sha1" 
                                        helperText="The last SHA1 commit id reviewed" 
                                        variant="outlined" 
                                        margin="dense" 
                                        error={editSha1?.length != 40 && editSha1?.length != 0}
                                        onInput={e=>setEditSha1(e.target.value)}
                                        value={editSha1 ? editSha1 : ""}
                                        disabled={editUrl == ""}
                                        InputLabelProps={{ shrink: shrink }}/>
                                </FormControl>
                            </Grid>
                        </Grid> 
                        <br/>
                        <Button 
                            variant="outlined" 
                            size="large"
                            onClick={handleEditRepo}
                            disabled={editUrl == ""}
                        >
                            Edit
                        </Button>
                    </form>   
                    <br/>
                </Grid>
            </Grid>   
            
            <form ref={formRef} onSubmit={saveConfig}>                
                <Button variant="outlined" type="submit" className="save_button" size="large">
                {
                    loading ? "Working ..." : "Save"
                }
                </Button>                
                <h3>Advanced Editor</h3>
                <TextareaAutosize
                    aria-label="empty textarea"
                    placeholder="Empty"
                    defaultValue={pretty}
                    className="config"
                    name="yaml_config"
                />
            </form>
        </>
    );
}

function ServerConfig() {
    const [gitconfig, setGitConfig] = useState<string>();
    const [sshkeys, setSSHKeys] = useState<string>();
    const [loading, setloadingResult] = useState<boolean>();
    const formRef = useRef()

    useEffect(() => {
        const get_git_config = async () => {
            const data = await fetch(server + "/gitconfig", {
                method: "GET"
            });
            const text = await data.text();
            setGitConfig(text);
        };
    
        get_git_config();

        const get_ssh_keys = async () => {
            const data = await fetch(server + "/sshkeys", {
                method: "GET"
            });
            const text = await data.text();
            setSSHKeys(text);
        };
    
        get_ssh_keys();
    }, []);

    const saveConfig = async(event: any) => {   
        setloadingResult(true);
        event.preventDefault();
        const data = new FormData(formRef.current)
        
        const data1 = await fetch(server + "/sshkeys", {
            method: "POST",
            mode: 'cors', 
            body: data
        });
        
        await data1.text();

        const data2 = await fetch(server + "/gitconfig", {
            method: "POST",
            mode: 'cors', 
            body: data
        });
        
        await data2.text();

        setloadingResult(false);
    }

    return (
        <form ref={formRef} onSubmit={saveConfig}>
            
            <Button variant="outlined" type="submit" className="save_button">
            {
                loading ? "Working ..." : "Save"
            }
            </Button>
            <h3>Configure .gitconfig</h3>
            <TextareaAutosize
                aria-label="empty textarea"
                placeholder="Empty"
                defaultValue={gitconfig}
                className="config"
                name="git_config"
            />
            <h3>Configure ssh keys</h3>
            <TextareaAutosize
                aria-label="empty textarea"
                placeholder="Empty"
                defaultValue={sshkeys}
                className="config"
                name="ssh_keys"
            />
        </form>
    );
}

function ConfigState() {
    const [stateConfig, setStateResult] = useState<string>();
    const [refreshInterval, setRefreshInterval] = useState(refresh_interval);

    const get_state = async () => {
        const data = await fetch(server + "/state", {
            method: "GET"
        });
        const text = await data.text();
        setStateResult(text);
    };

    useEffect(() => {
        if (refreshInterval && refreshInterval > 0) {
          const interval = setInterval(get_state, refreshInterval);
          return () => clearInterval(interval);
        }
      }, [refreshInterval]);

    useEffect(() => {  
        setRefreshInterval(refresh_interval)  
        get_state()
    }, []);

    return <><h3>Server state</h3><pre id="state_config" className="configState">{stateConfig}</pre></>;
}

export const Layout = () => {
    const [value, setValue] = React.useState(0);

    const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
        setValue(newValue);
    };

    const darkTheme = createTheme({
        palette: {
          mode: 'dark',
        }
    });

    return (
        <div className="main_wrapper">
            <pre className='logo'>{logo}</pre>
            <ThemeProvider theme={darkTheme}>
                <StyledEngineProvider injectFirst>
                    <Box sx={{ width: '100%' }}>
                        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                        <Tabs value={value} onChange={handleChange} aria-label="basic tabs example" centered>
                            <Tab label="Latest commits" {...a11yProps(0)} />
                            <Tab label="Repo Configuration" {...a11yProps(1)} />
                            <Tab label="Server Configuration" {...a11yProps(1)} />
                            <Tab label="Server State" {...a11yProps(2)} />
                            <Tab label="Server logs" {...a11yProps(3)} />
                        </Tabs>
                        </Box>
                        <TabPanel value={value} index={0}>
                            <Commits />
                        </TabPanel>
                        <TabPanel value={value} index={1}>
                            <Config />
                        </TabPanel>
                        <TabPanel value={value} index={2}>
                            <ServerConfig/>
                        </TabPanel>
                        <TabPanel value={value} index={3}>
                            <ConfigState />
                        </TabPanel>
                        <TabPanel value={value} index={4}>
                            <ServerLog/>
                        </TabPanel>
                    </Box>
                </StyledEngineProvider>
            </ThemeProvider>
        </div>
    );
}
