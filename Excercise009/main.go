package main

import (
	"fmt"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	glfw.Init()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	vk.Init()

	//1. Instance
	//	1. Create Application Struct
	//	2. Create Instance
	//	3. Create Window
	//	4. Create Surface
	var instance vk.Instance
	var appInfo vk.ApplicationInfo = vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.ApiVersion10,
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	var layers = []string{"VK_LAYER_NV_optimus\x00"}
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PNext:                   nil,
		Flags:                   0,
		PApplicationInfo:        &appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	result := vk.CreateInstance(&instanceInfo, nil, &instance)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating instance: %v", result))
	}
	var window *glfw.Window
	var err error
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err = glfw.CreateWindow(640, 480, "Vulkan Info", nil, nil)
	if err != nil {
		fmt.Println("Failed to create window with error :", err)
	}
	var surface vk.Surface
	pSurface, glfwError := window.CreateWindowSurface(instance, nil)
	if glfwError != nil {
		fmt.Println("Failed to create window surface with error :", glfwError)
	}
	surface = vk.SurfaceFromPointer(pSurface)
	fmt.Println(surface)
	//2. Logical Device
	//	1. Get All Physical Devices GPU
	//  2. List Physical Devices GPU properties
	//  3. Get MemoryProprties of Selected Physical Device
	//	4. Get Device Extension Properties
	//	5. Create Logical Device
	var deviceCount uint32
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	fmt.Printf("\tFound total %v Physical Device(s).....\n", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var physicalDevice = physicalDevices[0]
	var physicalDeviceMemoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physicalDevice, &physicalDeviceMemoryProperties)
	deviceQueueCreateInfo := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: 0,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	var deviceProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(physicalDevice, &deviceProperties)
	deviceProperties.Deref()
	fmt.Println("Driver Details...............")
	fmt.Println("\t* Driver:\t", deviceProperties.DriverVersion)
	fmt.Println("\t* Type:\t\t", deviceProperties.DeviceType)
	fmt.Printf("\t* Version: \t%v.%v.%v\n", (deviceProperties.ApiVersion>>22)&0x3FF, (deviceProperties.ApiVersion>>22)&0x3FF, deviceProperties.ApiVersion&0xFFF)
	fmt.Println("\t* Name:\t", string(deviceProperties.DeviceName[:]))
	var deviceExtensionCount uint32
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var deviceExtensions = make([]vk.ExtensionProperties, deviceExtensionCount)
	var deviceExtensionNames = make([]string, deviceExtensionCount)
	for idx, extension := range deviceExtensions {
		extension.Deref()
		deviceExtensionNames[idx] = fmt.Sprintf("%v\x00", extension.ExtensionName[:])
	}
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, deviceExtensions)
	var deviceFeatures = make([]vk.PhysicalDeviceFeatures, 1)
	deviceFeatures[0].ShaderClipDistance = vk.True
	var deviceCreateInfo vk.DeviceCreateInfo = vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(deviceQueueCreateInfo)),
		PQueueCreateInfos:       deviceQueueCreateInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(deviceExtensionNames)),
		PpEnabledExtensionNames: deviceExtensionNames,
		PEnabledFeatures:        deviceFeatures,
	}
	var logicalDevice vk.Device
	result = vk.CreateDevice(physicalDevice, &deviceCreateInfo, nil, &logicalDevice)

	//3. Create SwapChain
	//	1. Get surface capabilities
	//  2. Create Swapchain
	var physicalDeviceSurfaceCapabilities vk.SurfaceCapabilities
	result = vk.GetPhysicalDeviceSurfaceCapabilities(physicalDevice, surface, &physicalDeviceSurfaceCapabilities)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device surface capabilities: %v", result))
		panic(result)
	}
	var surfaceResuloution vk.Extent2D
	surfaceResuloution = physicalDeviceSurfaceCapabilities.CurrentExtent
	width := surfaceResuloution.Width
	height := surfaceResuloution.Height
	var swapChainInfo = vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          surface,
		MinImageCount:    2,
		ImageFormat:      vk.FormatB8g8r8a8Unorm,
		ImageColorSpace:  vk.ColorspaceSrgbNonlinear,
		ImageExtent:      surfaceResuloution,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		ImageSharingMode: vk.SharingModeExclusive,
		PreTransform:     vk.SurfaceTransformIdentityBit,
		CompositeAlpha:   vk.CompositeAlphaOpaqueBit,
		PresentMode:      vk.PresentModeMailbox,
		Clipped:          vk.True,
		OldSwapchain:     nil,
	}
	var swapChain vk.Swapchain
	result = vk.CreateSwapchain(logicalDevice, &swapChainInfo, nil, &swapChain)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating swapchain: %v", result))
		panic(result)
	}
	//4. Create Image and ImageView
	//	1. Get coount of images required by swapchain
	//  2. Create Swapchain
	var imageCount uint32
	result = vk.GetSwapchainImages(logicalDevice, swapChain, &imageCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting image count from swapchain: %v", result))
		panic(result)
	}
	fmt.Println("Querying Number of Images required by swapchain....", imageCount)
	var images []vk.Image
	images = make([]vk.Image, imageCount)
	result = vk.GetSwapchainImages(logicalDevice, swapChain, &imageCount, images)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting images for swapchain: %v", result))
		panic(result)
	}
	var imageViews = make([]vk.ImageView, 2)
	for idx := range imageViews {
		var imageViewCreateInfo = vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			ViewType: vk.ImageViewType2d,
			Format:   vk.FormatB8g8r8a8Unorm,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			Image: images[idx],
		}
		result = vk.CreateImageView(logicalDevice, &imageViewCreateInfo, nil, &imageViews[idx])
		if result != vk.Success {
			fmt.Println(fmt.Errorf("Failed to create image view : %v", result))
			panic(result)
		}
	}
	//5. Create and Allocate CommandBuffer
	//	1. Get first Queue
	//  2. Create commmand pool
	//  3. Allocate commmand buffer
	var queue vk.Queue
	vk.GetDeviceQueue(logicalDevice, 0, 0, &queue)
	var commandPool vk.CommandPool
	var commandPoolCreateInfo = vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeQueryPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	result = vk.CreateCommandPool(logicalDevice, &commandPoolCreateInfo, nil, &commandPool)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to create command pool : %v", result))
		panic(result)
	}
	var commandBuffer []vk.CommandBuffer
	commandBuffer = make([]vk.CommandBuffer, 1)
	var commandBufferAllocateInfo = vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	result = vk.AllocateCommandBuffers(logicalDevice, &commandBufferAllocateInfo, commandBuffer)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to allocate command buffer : %v", result))
		panic(result)
	}
	//6. Create FrameBuffer
	//	1. Create Attachment Description and Attachment Reference
	//  2. Create Subpass
	//  3. Create Render Pass
	//  4. Create FrameBuffer
	var attachmentDescriptions []vk.AttachmentDescription
	attachmentDescriptions = make([]vk.AttachmentDescription, 1)
	attachmentDescriptions[0] = vk.AttachmentDescription{
		Format:         vk.FormatB8g8r8a8Unorm,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutColorAttachmentOptimal,
		FinalLayout:    vk.ImageLayoutColorAttachmentOptimal,
	}
	var attachmentReference = make([]vk.AttachmentReference, 1)
	attachmentReference[0].Attachment = 0
	attachmentReference[0].Layout = vk.ImageLayoutColorAttachmentOptimal
	var subpass = make([]vk.SubpassDescription, 1)
	subpass[0] = vk.SubpassDescription{
		PipelineBindPoint:       vk.PipelineBindPointGraphics,
		ColorAttachmentCount:    1,
		PColorAttachments:       attachmentReference,
		PDepthStencilAttachment: nil,
	}
	var renderPassCreateInfo = vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    attachmentDescriptions,
		SubpassCount:    1,
		PSubpasses:      subpass,
	}
	var renderPass vk.RenderPass
	result = vk.CreateRenderPass(logicalDevice, &renderPassCreateInfo, nil, &renderPass)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Failed to create render pass : %v", result))
		panic(result)
	}
	var frameBufferAttachments = make([]vk.ImageView, 1)
	var frameBufferCreateInfo = vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		PNext:           nil,
		RenderPass:      renderPass,
		AttachmentCount: 1,
		PAttachments:    frameBufferAttachments,
		Width:           width,
		Height:          height,
		Layers:          1,
	}
	var frameBuffers = make([]vk.Framebuffer, 2)
	for idx := range frameBuffers {
		result = vk.CreateFramebuffer(logicalDevice, &frameBufferCreateInfo, nil, &frameBuffers[idx])
		if result != vk.Success {
			fmt.Println(fmt.Errorf("Failed to create frame buffers : %v", result))
			panic(result)
		}
	}
	//Cleanup
	for idx := range frameBuffers {
		vk.DestroyFramebuffer(logicalDevice, frameBuffers[idx], nil)
	}
	vk.DestroyRenderPass(logicalDevice, renderPass, nil)
	vk.DestroyCommandPool(logicalDevice, commandPool, nil)
	for _, imageView := range imageViews {
		vk.DestroyImageView(logicalDevice, imageView, nil)
	}
	for _, image := range images {
		vk.DestroyImage(logicalDevice, image, nil)
	}
	vk.DestroySwapchain(logicalDevice, swapChain, nil)
	vk.DestroySurface(instance, surface, nil)
	vk.DestroyInstance(instance, nil)
	window.Destroy()
}
